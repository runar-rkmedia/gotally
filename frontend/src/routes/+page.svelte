<script lang="ts">
	export const ssr = false
	import 'pollen-css'
	import {
		GameMode,
		GetHintRequest,
		httpErrorStore,
		Indexes,
		SwipeDirection,
		UndoRequest
	} from '../connect-web'
	import { onMount } from 'svelte'
	import { browser } from '$app/environment'
	import { animateSwipe, coordToIndex, createSelectionDirectionMap, ValidatePath } from '../logic'
	import type { PartialMessage } from '@bufbuild/protobuf/dist/types/message'
	import { ErrNoChange, store, storeHandler } from '../connect-web/store'
	import SwipeHint from '../components/board/SwipeHint.svelte'
	import GameWon from '../components/GameWon.svelte'
	import GameMenu from '../components/GameMenu.svelte'
	import Dialog from '../components/Dialog.svelte'
	import CellComp from '../components/board/Cell.svelte'
	import { cellValue } from '../components/board/cell'
	import Counter from '../components/Counter.svelte'
	import userSettings from '../userSettings'
	import PrimeFactors from '../components/PrimeFactors.svelte'
	import { findCellFromTouch } from '../utils/touchHandlers'
	import Icon from '../components/Icon.svelte'
	import { findDOMParent } from '../utils/findDomParent'

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
	const undo = async (options?: PartialMessage<UndoRequest>) => {
		console.log('undo?', storeHandler)
		return storeHandler.commit(storeHandler.undo(options))
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
			console.log('drag reset', diff)
			return true
		}
		console.log('drag reset NOT', diff)
		return false
	}
	const swipeQueue: SwipeDirection[] = []
	const swipe = async (direction: SwipeDirection) => {
		if (selection.length) {
			if (!didDrag) {
				return
			}
			if (swipeLockedForDragging()) {
				resetSelection()
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
				resetSelection()
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
			nRows: $store.session.game.board.rows
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
	function createSwiper(node: HTMLElement) {
		if (!browser) {
			return
		}
		let Hammer: any
		import('hammerjs').then((h) => {
			Hammer = h.default
			const hammerTime = new Hammer(node, {
				recognizers: [[Hammer.Swipe, { direction: Hammer.DIRECTION_ALL }]]
			})
			hammerTime.on('swipe', (e: any) => {
				switch (e.direction) {
					case Hammer.DIRECTION_UP:
						swipe(SwipeDirection.UP)
						break
					case Hammer.DIRECTION_DOWN:
						swipe(SwipeDirection.DOWN)
						break
					case Hammer.DIRECTION_LEFT:
						swipe(SwipeDirection.LEFT)
						break
					case Hammer.DIRECTION_RIGHT:
						swipe(SwipeDirection.RIGHT)
						break
				}
			})
		})
	}
	let selection: number[] = []
	let lastSelectionValue = 0
	let selectionMap: Record<number, boolean | undefined> = {}
	let invalidSelectionMap: Record<number, boolean | undefined> = {}
	// let selectionDirectionMap: Record<number, pathDirection> = {
	// 	11: 'up',
	// 	6: 'continue',
	// 	1: 'upright',
	// 	2: 'right'
	// }
	$: selectionDirectionMap = createSelectionDirectionMap(selection)
	let pathInvalidErr: any
	const showSelectionInfo = true
	const resetSelection = () => {
		invalidSelectionMap = {}
		selection = []
		lastSelectionValue = 0
		selectionMap = {}
	}
	const select = async (i: number) => {
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
		selectionMap[i] = true
	}
	$: nextHint = $store.hintDoneIndex >= 0 ? $store.hints[$store.hintDoneIndex + 1] : $store.hints[0]
	$: selectionSum = !showSelectionInfo
		? 0
		: selection.reduce((r, i) => r + cellValue($store.session.game.board.cells[i]), 0)
	$: selectionEvaluatedSum = !showSelectionInfo
		? 0
		: selection.slice(0, -1).reduce((r, i) => r + cellValue($store.session.game.board.cells[i]), 0)
	$: selectionProduct = !showSelectionInfo
		? 0
		: selection.reduce((r, i) => r * cellValue($store.session.game.board.cells[i]), 1)
	$: selectionEvaluatedProduct = !showSelectionInfo
		? 0
		: selection.slice(0, -1).reduce((r, i) => r * cellValue($store.session.game.board.cells[i]), 1)

	$: pathEvaluatesToLast =
		selection.length >= 2 &&
		(lastSelectionValue === selectionEvaluatedSum ||
			lastSelectionValue === selectionEvaluatedProduct)
	$: {
		console.log('eval', lastSelectionValue, selectionEvaluatedSum, selectionSum, selection)
	}
	$: {
		console.log('mouseDown state', mouseDown)
	}
	let didDrag: Date | null = null
	let canDragToSelect = true
	let canSelectNonNeighbours = false
	let resetSelectionOnSwipe = true
	let boardWidth: number = 682
	let boardHeight: number = 500
	$: boardCellWidth = (boardWidth || 100) / ($store?.session?.game?.board?.columns || 1)
	$: boardCellHeight = (boardHeight || 100) / ($store?.session?.game?.board?.columns || 1)
	$: boardCellSize = Math.min(boardCellWidth, boardCellHeight)
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
		<div class="headControls">
			<div class="info">
				<div>
					<div class="score" data-score={$store.session.game.score} data-testid="score">
						Score: {$store.session.game.score}
					</div>
					<small class="boardName" title={$store.session.game.board.id}
						>{$store.session.game.board.name}</small
					>
					<div class="moves" data-moves={$store.session.game.moves}>
						Moves: {$store.session.game.moves}
					</div>
				</div>
			</div>
			<p>Hi, {$store.session.username} ({$store.session.sessionId})</p>
			<button
				class="icon-only"
				data-testid="menu"
				on:click={() => (showGameMenu = true)}
				aria-roledescription="Show menu"
			>
				<Icon icon="settings" color="white" />
			</button>
		</div>

		<div class="boardContainer">
			<SwipeHint
				instruction={nextHint?.instructionOneof.value}
				active={nextHint?.instructionOneof.case === 'swipe'}
			/>
			<div
				bind:this={boardDiv}
				bind:clientWidth={boardWidth}
				bind:clientHeight={boardHeight}
				use:createSwiper
				on:touchend|preventDefault={(e) => {
					if (isSwiping) {
						return
					}
					if (!canDragToSelect) {
						console.log('touchend no-can-drag')
						return
					}
					if (!didDrag) {
						console.log('touchend no-drag')
						return
					}
					didDrag = null
					const [findings, err] = findCellFromTouch(e)
					if (err) {
						console.error(err.message, err.details)
						resetSelection()
						return
					}
					// if (selection[selection.length - 1] === findings.index) {
					// 	console.log('touchend last')
					// 	return
					// }
					if (selection.length === 1 && selection[0] === findings.index) {
						return
					}
					console.log('touchend-select')
					select(findings.index)
				}}
				on:touchmove|preventDefault={(e) => {
					if (!canDragToSelect) {
						return
					}
					if (isSwiping) {
						return
					}
					const [findings, err] = findCellFromTouch(e)
					if (err) {
						// console.error(err.message, err.details)
						return
					}
					if (selectionMap[findings.index]) {
						return
					}
					select(findings.index)
					didDrag = new Date()
				}}
				on:touchstart|preventDefault={(e) => {
					if (isSwiping) {
						return
					}
					if (!canDragToSelect) {
						return
					}
					const [findings, err] = findCellFromTouch(e)
					if (err) {
						resetSelection()
						return
					}
					if (selectionMap[findings.index]) {
						if (selection.length === 1) {
							resetSelection()
							return
						}
						if (selection[selection.length - 1] !== findings.index) {
							resetSelection()
							return
						}
						// combine
						select(findings.index)
						return
					}
					select(findings.index)
				}}
				class="board"
				style={`
grid-template-columns: repeat(${$store.session.game.board.columns}, 1fr); 
grid-template-rows: repeat(${$store.session.game.board.rows}, 1fr);
        --board-cell-width: ${boardCellSize}px;
        --board-cell-height: ${boardCellSize}px;
`}
			>
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
							e.preventDefault()
							if (e.ctrlKey) {
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
							if (!didDrag) {
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
						on:mouseenter={(e) => {
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
							select(i)
						}}
						on:mousedown={() => {
							select(i)
							didDrag = null
						}}
					/>
				{/each}
			</div>
		</div>
		{#if showSelectionInfo}
			<div class="selectionCounter">
				<Counter
					show={!!selectionSum}
					asCell={false}
					value={selectionSum}
					label="Sum"
					variant={lastSelectionValue * 2 < selectionSum
						? 'error'
						: lastSelectionValue * 2 === selectionSum
						? 'success'
						: 'normal'}
				/>
				<Counter
					show={selectionProduct > 1}
					asCell={true}
					value={selectionProduct}
					label="Product"
					variant={lastSelectionValue < selectionProduct / lastSelectionValue
						? 'error'
						: lastSelectionValue === selectionProduct / lastSelectionValue
						? 'success'
						: 'normal'}
				/>
				<PrimeFactors n={selectionProduct} />
			</div>
		{/if}
		<p>
			{$store.session.game.description}
		</p>
		<div class="bottom-controls">
			<button
				data-testid="undo"
				on:click={() => undo()}
				disabled={$store.didWin || $store.session.game.moves <= 0}
			>
				<Icon icon="undo" color="white" /> Undo
			</button>
			<button data-testid="hint" on:click={() => getHint()} disabled={$store.didWin}>
				<Icon icon="help" color="white" /> Hint
			</button>
		</div>
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
	.board {
		position: relative;
		transition: opacity 300ms var(--easing-standard);
		/* margin-inline: -4px; */
		display: grid;
		height: 100%;
	}
	.boardName {
		opacity: 0.7;
		float: right;
	}
	.selectionCounter {
		display: grid;
		grid-template-columns: 5fr 5fr 2fr;
		gap: 10px;
	}
	.bottom-controls {
		display: flex;
		justify-content: center;
		flex-direction: column;
	}
	button:disabled {
		opacity: 0.4;
	}
	button {
		cursor: pointer;
		transition: opacity 70ms var(--easing-standard);
		min-width: 52px;
		min-height: 52px;
		color: var(--color-white);
		&:not(icon-only) {
			background-color: var(--color-primary);
		}

		&.icon-only {
			min-height: 48px;
			background: unset;
			border: unset;
			display: flex;
			justify-content: center;
			align-items: center;
		}
	}
	.headControls {
		display: flex;
		justify-content: space-between;
	}
</style>
