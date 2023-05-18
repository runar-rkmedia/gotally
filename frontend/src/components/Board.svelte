<script lang="ts">
	import { browser } from '$app/environment'
	import { SwipeDirection } from '../connect-web'
	import { store } from '../connect-web/store'
	import { findCellFromTouch } from '../utils/touchHandlers'

	export let swipe: (direction: SwipeDirection) => void
	export let resetSelection: () => void
	export let select: (i: number) => void
	export let boardDiv: HTMLDivElement
	export let selection: number[]
	export let selectionMap: Record<number, any>
	export let isSwiping: boolean
	export let canDragToSelect: boolean = true
	export let didDrag: Date | null = null

	let boardWidth: number = 682
	let boardHeight: number = 500

	$: boardCellWidth = (boardWidth || 100) / ($store?.session?.game?.board?.columns || 1)
	$: boardCellHeight = (boardHeight || 100) / ($store?.session?.game?.board?.columns || 1)
	$: boardCellSize = Math.min(boardCellWidth, boardCellHeight)
	function createSwiper(node: HTMLElement) {
		if (!browser) {
			return
		}
		let Hammer: any
		import('hammerjs').then((h) => {
			Hammer = h.default
			const hammerTime = new Hammer(node, {
				recognizers: [[Hammer.Swipe, { direction: Hammer.DIRECTION_ALL }]],
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
</script>

<div
	bind:this={boardDiv}
	bind:clientWidth={boardWidth}
	bind:clientHeight={boardHeight}
	use:createSwiper
	on:touchend|preventDefault={(e) => {
		console.log('touchend')
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
	<slot />
</div>

<style lang="scss">
	.board {
		position: relative;
		transition: opacity 300ms var(--easing-standard);
		/* margin-inline: -4px; */
		display: grid;
		height: 100%;
	}
</style>
