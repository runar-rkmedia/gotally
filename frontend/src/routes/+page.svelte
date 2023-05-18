<script lang="ts">
	export const ssr = false
	import 'pollen-css'
	import { GameMode, GetHintRequest, SwipeDirection } from '../connect-web'
	import { onMount } from 'svelte'
	import { browser } from '$app/environment'
	import { animateSwipe, coordToIndex, createSelectionDirectionMap, ValidatePath } from '../logic'
	import type { PartialMessage } from '@bufbuild/protobuf/dist/types/message'
	import { ErrNoChange, store, storeHandler } from '../connect-web/store'
	import SwipeHint from '../components/board/SwipeHint.svelte'
	import Board from '../components/Board.svelte'
	import GameWon from '../components/GameWon.svelte'
	import GameMenu from '../components/GameMenu.svelte'
	import Dialog from '../components/Dialog.svelte'
	import CellComp from '../components/board/Cell.svelte'
	import { cellValue } from '../components/board/cell'
	import userSettings from '../userSettings'
	import { findDOMParent } from '../utils/findDomParent'
	import GameHeader from '../components/GameHeader.svelte'
	import GameButtons from '../components/GameButtons.svelte'
	import SelectionInfo from '../components/SelectionInfo.svelte'

	$: {
		// when user wins a game, refresh the challenge
		if ($store.didWin) {
			console.log('refreshing challenges')
			storeHandler.commit(storeHandler.getChallenges({}))
		}
	}
	let boardDiv: HTMLDivElement
	let showGameMenu = false

	const getHint = async (options?: PartialMessage<GetHintRequest>) => {
		return storeHandler.commit(storeHandler.getHint(options))
	}
	const undo = async () => {
		return storeHandler.commit(storeHandler.undo())
	}
	let lastNumberKey: number | null = null

	const selectByNumber = (n: number) => {
		// Threat as if the coordinate-system starts at 1
		n = n - 1
		if (n < 0) {
			lastNumberKey = null
			return
		}
		if (lastNumberKey !== null) {
			const index = coordToIndex(
				lastNumberKey,
				n,
				$store.session.game.board.columns,
				$store.session.game.board.rows
			)
			if (index !== null) {
				select(index)
			}
			lastNumberKey = null
			return
		}
		lastNumberKey = n
	}
	let mouseDown = false

	onMount(async () => {
		await storeHandler.commit(storeHandler.getSession())

		// Set up swipes
		if (browser) {
			document.onmousedown = () => (mouseDown = true)
			document.onmouseup = () => {
				if (didDrag) {
					setTimeout(() => {
						if (!didDrag) {
							return
						}
						resetSelection()
					}, 200)
				}
				mouseDown = false
			}
			document.onkeyup = async (e) => {
				if ($store.didWin) {
					return
				}
				if (swipeLock) {
					e.preventDefault()
					return
				}
				if (e.ctrlKey) {
					return
				}
				if (e.shiftKey) {
					return
				}
				if (e.altKey || e.metaKey) {
					return
				}
				if (
					document.activeElement &&
					(document.activeElement?.tagName === 'INPUT' ||
						document.activeElement?.tagName === 'textarea' ||
						findDOMParent(document.activeElement, (el) => el.tagName === 'FORM'))
				) {
					return
				}
				const hasSelection = selection.length
				const lastSelection = selection[selection.length - 1]
				switch (e.key) {
					case '1':
					case '2':
					case '3':
					case '4':
					case '5':
					case '6':
					case '7':
					case '8':
					case '9':
					case '0':
						selectByNumber(Number(e.key))
						return
					case 'Escape':
						if (showGameMenu) {
							showGameMenu = false
							return
						}
						if (hasSelection) {
							selection = []
							selectionMap = {}
						}
						break
					case 'm':
						showGameMenu = !showGameMenu
						break
					case 'ArrowLeft':
					case 'a':
						if (hasSelection) {
							const next = lastSelection - 1
							if (lastSelection % $store.session?.game.board.columns! === 0) {
								return
							}
							if (selection.includes(next)) {
								return
							}
							select(next)
							return
						}
						swipe(SwipeDirection.LEFT)
						break
					case 'ArrowRight':
					case 'd':
						if (showGameMenu) {
							return
						}
						if (hasSelection) {
							const next = lastSelection + 1
							if (next % $store.session?.game.board.columns === 0) {
								return
							}
							if (selection.includes(next)) {
								return
							}
							select(lastSelection + 1)
							return
						}
						swipe(SwipeDirection.RIGHT)
						break
					case 'ArrowDown':
					case 's':
						if (showGameMenu) {
							return
						}
						if (hasSelection) {
							const next = lastSelection + $store.session.game.board.rows
							if (next > $store.session.game.board.rows * $store.session.game.board.columns) {
								return
							}
							if (selection.includes(next)) {
								return
							}
							select(next)
							return
						}
						swipe(SwipeDirection.DOWN)
						break
					case 'ArrowUp':
					case 'w':
						if (showGameMenu) {
							return
						}
						if (hasSelection) {
							const next = lastSelection - $store.session.game.board.rows
							if (next < 0) {
								return
							}
							if (selection.includes(next)) {
								return
							}
							select(next)
							return
						}
						swipe(SwipeDirection.UP)
						break
					case 'h':
						if (showGameMenu) {
							return
						}
						getHint()
						break
					case 'u':
						if (showGameMenu) {
							return
						}
						undo()
						break
					// Combine path
					case 'Space':
					case ' ':
						if (showGameMenu) {
							return
						}
						if (!hasSelection) {
							return
						}
						select(lastSelection)
						break
					case 'r':
						return
						if (!$store?.session.game.moves) {
							return
						}
						storeHandler.commit(storeHandler.restartGame())
						break
					case 'n':
						storeHandler.commit(storeHandler.newGame({ mode: GameMode.RANDOM_CHALLENGE }))
						break

					default:
						return
				}
				e.preventDefault()
			}
		}
	})
	const animateInvalidSwipe = (direction: SwipeDirection) => {
		console.error('Not implemented', 'animateInvalidSwipe', direction)
	}
	let swipeLock = false
	let isSwiping = false
	let _swipeQueueHandling = false
	const swipeLockedForDragging = () => {
		if (!didDrag) {
			return false
		}
		// block swiping if dragging across cells
		const diff = new Date().getTime() - didDrag.getTime()
		if (diff > 200) {
			console.log('drag lock reset', diff)
			return true
		}
		console.log('drag lock reset NOT', diff)
		return false
	}
	const swipeQueue: SwipeDirection[] = []
	const swipe = async (direction: SwipeDirection) => {
		if (selection.length) {
			if (!didDrag) {
				return
			}
			if (swipeLockedForDragging()) {
				// resetSelection()
				return
			}
		}
		swipeQueue.push(direction)
		if (_swipeQueueHandling) {
			return
		}
		_swipeQueueHandling = true
		while (swipeQueue.length) {
			const dir = swipeQueue.pop()
			if (!dir) {
				_swipeQueueHandling = false
				return
			}
			await _swipe(dir)
		}
		_swipeQueueHandling = false
	}

	const _swipe = async (direction: SwipeDirection) => {
		if (!$store?.session?.game?.board) {
			return
		}
		if (selection.length) {
			if (!didDrag) {
				return
			}
			// block swiping if dragging across cells
			if (swipeLockedForDragging()) {
				// resetSelection()
				return
			}
		}
		if (swipeLock) {
			return
		}
		const swipeOptions = {
			swipeAnimationTime: $userSettings.swipeAnimationTime,
			positive: direction === SwipeDirection.DOWN || direction === SwipeDirection.RIGHT,
			vertical: direction === SwipeDirection.UP || direction === SwipeDirection.DOWN,
			boardEl: boardDiv,
			nColumns: $store.session.game.board.columns,
			nRows: $store.session.game.board.rows,
		}
		const shouldAnimate = await animateSwipe({ ...swipeOptions, dry: true })
		if (!shouldAnimate) {
			return
		}
		// swipeLock = true
		isSwiping = true
		resetSelection()
		const r = storeHandler.swipe(direction)
		await animateSwipe(swipeOptions)
		// In case the server responds slower than the animation,
		// we set a visual indicator here.
		// On the other hand, if the server is faster, the should should see nothing of this
		boardDiv.style.opacity = '0.8'
		const [_, commit, err] = await r
		swipeLock = false
		isSwiping = false
		for (const cell of [...boardDiv.children] as HTMLElement[]) {
			cell.style.transform = ''
			cell.style.transition = 'none'
		}
		boardDiv.style.opacity = '1'
		if (err) {
			if (err === ErrNoChange) {
				animateInvalidSwipe(direction)
			}
			return
		}
		commit()
	}
	let selection: number[] = []
	let lastSelectionValue = 0
	let selectionMap: Record<number, boolean | undefined> = {}
	let invalidSelectionMap: Record<number, boolean | undefined> = {}
	$: selectionDirectionMap = createSelectionDirectionMap(selection)
	let pathInvalidErr: any
	const resetSelection = () => {
		const err = new Error('')
		console.log('resetting selection', err.stack)
		invalidSelectionMap = {}
		selection = []
		lastSelectionValue = 0
		selectionMap = {}
	}
	const select = async (i: number) => {
		console.log('select', i, selection)
		invalidSelectionMap = {}
		const cell = $store.session.game.board.cells[i]
		if (!cell?.base) {
			resetSelection()
			return
		}
		const isSelected = !!selectionMap[i]
		pathInvalidErr = ValidatePath(
			isSelected ? selection : [...selection, i],
			$store.session.game.board.rows,
			$store.session.game.board.columns,
			$store.session.game.board.cells
		)
		if (isSelected || (!canSelectNonNeighbours && pathInvalidErr?.code === 'non-neighbours')) {
			if (pathInvalidErr) {
				resetSelection()
				invalidSelectionMap = { [i]: true }
				return
			}
			const [_, commit, err] = await storeHandler.combineCells(selection)
			if (err) {
				resetSelection()
				invalidSelectionMap = { [i]: true }
				return
			}
			commit()
			resetSelection()
			return
		}
		selection = [...selection, i]
		lastSelectionValue = cellValue(cell)
		selectionMap = { ...selectionMap, [i]: true }
	}
	$: nextHint = $store.hintDoneIndex >= 0 ? $store.hints[$store.hintDoneIndex + 1] : $store.hints[0]
	$: selectionSum = selection.reduce((r, i) => r + cellValue($store.session.game.board.cells[i]), 0)
	$: selectionEvaluatedSum = selection
		.slice(0, -1)
		.reduce((r, i) => r + cellValue($store.session.game.board.cells[i]), 0)
	$: selectionProduct = selection.reduce(
		(r, i) => r * cellValue($store.session.game.board.cells[i]),
		1
	)
	$: selectionEvaluatedProduct = selection
		.slice(0, -1)
		.reduce((r, i) => r * cellValue($store.session.game.board.cells[i]), 1)

	$: pathEvaluatesToLast =
		selection.length >= 2 &&
		(lastSelectionValue === selectionEvaluatedSum ||
			lastSelectionValue === selectionEvaluatedProduct)
	$: {
		console.log('eval', lastSelectionValue, selectionEvaluatedSum, selectionSum, selection)
	}
	$: {
		console.log('mouseDown state', mouseDown, selection)
	}
	let didDrag: Date | null = null
	let canSelectNonNeighbours = false
</script>

{#if $store?.session?.game?.board}
	<Dialog bind:open={$store.didWin} let:open>
		<GameWon {open} />
	</Dialog>
	<Dialog bind:open={showGameMenu}>
		<GameMenu bind:open={showGameMenu} />
	</Dialog>
{/if}
<div class="gameView">
	{#if $store?.session?.game?.board}
		<GameHeader bind:showGameMenu />

		<div class="boardContainer">
			{#if nextHint?.instructionOneof.case === 'swipe'}
				<SwipeHint
					instruction={nextHint?.instructionOneof.value}
					active={nextHint?.instructionOneof.case === 'swipe'}
				/>
			{/if}
			<Board {select} {selection} {selectionMap} {resetSelection} {isSwiping} bind:boardDiv {swipe}>
				{#each $store.session.game.board.cells as c, i}
					<CellComp
						pathDir={selectionDirectionMap[i]}
						noEval={invalidSelectionMap[i]}
						selected={selectionMap[i]}
						hasHint={nextHint?.instructionOneof.case === 'combine'}
						hinted={nextHint?.instructionOneof.case === 'combine' &&
							nextHint.instructionOneof.value.index.includes(i)}
						hintedLast={nextHint?.instructionOneof.case === 'combine' &&
							nextHint.instructionOneof.value.index[
								nextHint.instructionOneof.value.index.length - 1
							] === i}
						selectedLast={!!selection.length && selection[selection.length - 1] === i}
						selectedFirst={!!selection.length && selection[0] === i}
						evaluatesTo={selection.length >= 2 && pathEvaluatesToLast}
						cell={c}
						on:mouseup={(e) => {
							console.log('mouseup')
							e.preventDefault()
							if (e.ctrlKey) {
								// This is only used in the generator / board-builder
								resetSelection()

								const val = cellValue(c)
								const res = prompt('Change this value', String(val))
								console.log('cccc', res, typeof res)
								if (res === null) {
									return
								}
								const n = Number(res)
								if (isNaN(n)) {
									alert('Must be a number')
									return
								}
								if (val === n) {
									return
								}
								$store.session.game.board.cells[i] = { base: n, twopow: 0 }

								return
							}
							console.log(
								'mouse up  select',
								i,
								didDrag,
								invalidSelectionMap[i],
								selectionMap[i],
								selection
							)
							if (!didDrag) {
								return
							}
							if (true) {
								select(i)
								return
							}
							if (invalidSelectionMap[i]) {
								didDrag = null
								return
							}
							if (!selectionMap[i]) {
								didDrag = null
								return
							}
							select(i)
							didDrag = null
						}}
						on:mouseenter={() => {
							if (!mouseDown) {
								return
							}
							if (isSwiping) {
								return
							}
							if (selectionMap[i]) {
								return
							}
							if (!didDrag) {
								didDrag = new Date()
							}
							console.log('mouseneter select', i)
							select(i)
						}}
						on:mousedown={() => {
							console.log('mouse down ')
							select(i)
							didDrag = null
						}}
					/>
				{/each}
			</Board>
		</div>
		<SelectionInfo {selectionProduct} {selectionSum} {lastSelectionValue} />

		<p>
			{$store.session.game.description}
		</p>
		<GameButtons
			on:undo={() => undo()}
			on:hint={() => getHint()}
			didWin={$store.didWin}
			moves={$store.session.game.moves}
		/>
	{/if}
</div>

<style lang="scss">
	p {
		padding-inline: var(--size-4);
		padding-block-end: var(--size-2);
	}
	.gameView {
		height: 100%;
		max-height: 100%;
		display: flex;
		flex-direction: column;
	}

	.boardContainer {
		position: relative;
		border: 2px solid var(--color-blue-700);
		border-radius: var(--radius-lg);
		margin-block-end: var(--size-4);
		height: 100%;
		max-height: 100%;
	}
</style>
