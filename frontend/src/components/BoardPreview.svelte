<script lang="ts">
	import { store } from '../connect-web/store'
	import CellComp from './board/Cell.svelte'
	export let cells: { base: number; twopow: number }[]
	export let rows: number
	export let columns: number
	export let heightRatio: number = 9 / 16
	let clientWidth: number

	// visibility-observer
	import { createEventDispatcher, onMount } from 'svelte'

	export let top = 0
	export let bottom = 0
	export let left = 0
	export let right = 0

	export let steps = 100

	let element: HTMLDivElement
	let percent: number
	let observer: IntersectionObserver
	let unobserve = () => {}
	let intersectionObserverSupport = false

	const intersectPercent: IntersectionObserverCallback = (entries) => {
		entries.forEach((entry) => {
			percent = Math.round(Math.ceil(entry.intersectionRatio * 100))
		})
	}

	function stepsToThreshold(steps: number) {
		return [...Array(steps).keys()].map((n) => n / steps)
	}

	onMount(() => {
		intersectionObserverSupport =
			'IntersectionObserver' in window &&
			'IntersectionObserverEntry' in window &&
			'intersectionRatio' in window.IntersectionObserverEntry.prototype

		const options = {
			rootMargin: `${top}px ${right}px ${bottom}px ${left}px`,
			threshold: stepsToThreshold(steps),
		}

		if (intersectionObserverSupport) {
			observer = new IntersectionObserver(intersectPercent, options)
			observer.observe(element)
			unobserve = () => observer.unobserve(element)
		}

		return unobserve
	})
	//
	const dispatch = createEventDispatcher<{
		cellclick: { cell: { base: number; twopow: number }; i: number }
	}>()
</script>

<div
	bind:this={element}
	class:visible={percent > 90}
	bind:clientWidth
	style={`
height: ${clientWidth * heightRatio}px;
--board-cell-width: ${clientWidth / columns}px;
--board-cell-height: ${(clientWidth * heightRatio) / rows}px;
grid-template-columns: repeat(${columns}, 1fr); grid-template-rows: repeat(${rows}, 1fr)`}
>
	{#each cells as c, i}
		<CellComp on:mouseup={(e) => dispatch('cellclick', { cell: c, i })} cell={c} />
	{/each}
</div>

<style lang="scss">
	div {
		display: grid;
		transform-style: preserve-3d;
		border-radius: 32px;
		outline: 4px solid var(--color-grey-900);
		background: var(--color-grey-900);
		overflow: hidden;
		transform: perspective(600px) rotateY(25deg) scale(0.9) rotateX(10deg);
		filter: blur(1px);
		opacity: 0.5;
		transition: 0.6s var(--easing-standard);
		&.visible,
		&:hover {
			transform: perspective(600px) rotateY(-15deg) translateY(0px) rotateX(10deg) scale(0.9);
			filter: blur(0);
			opacity: 1;
		}
	}
</style>
