<script lang="ts">
	import { numberFormatter } from '../formatNumber'

	import { cellValue, primeFactors } from './cell'

	// noEval={invalidSelectionMap[i]}
	// selected={selectionMap[i]}
	// hinted={nextHint?.instructionOneof.case === 'combine' &&
	// 	nextHint.instructionOneof.value.index.includes(i)}
	// selectedLast={!!selection.length && selection[selection.length - 1] === i}
	// base={c.base}
	type Cell = {
		base: number
		twopow: number
	}

	export let noEval: boolean | undefined
	export let selected: boolean | undefined
	export let hinted: boolean | undefined
	export let selectedLast: boolean | undefined
	export let evaluatesTo: boolean | undefined
	export let cell: Cell
	$: factors = primeFactors(cellValue(cell))
	$: angle = 100 / factors.length
	let h: number
	let w: number
	$: formattedValue = numberFormatter(cellValue(cell))
</script>

<div
	class="cell"
	class:no-eval={noEval}
	class:selected
	class:hinted
	class:evaluatesTo
	class:selectedLast
	class:blank={Number(cell.base) === 0}
	data-base={cell.base}
	on:click
	bind:clientHeight={h}
	bind:clientWidth={w}
	style={`--cell-width: ${w}px; --cell-height: ${h}px; --value-length: ${
		w / formattedValue.length
	}px`}
>
	<div class="factors" data-factors={factors.length}>
		<div class="inner">
			{#each factors as f, i}
				<div class="sector-wrapper">
					<div
						class="sector no-round"
						style={`--p: ${angle - (factors.length >= 2 ? 2 : 0)}; --o:  ${
							(360 / factors.length) * i + (factors.length >= 2 ? 4 : 0)
						}deg`}
						data-factor={f}
					/>
				</div>
			{/each}
		</div>
	</div>
	<div class="cellValue">
		<div class="inner">
			{formattedValue}
		</div>
	</div>
</div>

<style>
	.cell {
		--max-font-size: 1.8rem;
		--min-font-size: 1rem;
		--factor-width: 18vw;
		--factor-width: calc(min(var(--cell-height), var(--cell-width)) - 5px);
		--width-inner-circle: calc(var(--factor-width) * 0.8);
		--sector-width: var(--factor-width);
		--inner-padding: calc(var(--width-inner-circle) / 2 / 1.5);
		--ctb: var(--color-grey-50);
		--ctc: var(--color-black);
		font-weight: bold;
		user-select: none;
		display: flex;
		justify-content: center;
		align-items: center;
		border: 2px solid var(--border-blue-700);
		margin: 2px;
		border-radius: 8px;
		background-color: var(--color-grey-700);
		position: relative;
		color: var(--ctc);
	}
	.cell:empty,
	.cell.blank {
		opacity: 0;
		visibility: hidden;
	}
	.cell.hinted:not(.selected) {
		background-color: var(--color-blue-500);
		outline-color: var(--color-purple-700);
		outline-width: 5px;
		outline-style: dotted;
	}
	.cell::before {
		position: absolute;
		content: '';
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;

		opacity: 0.3;
	}
	.factors {
		background-color: white;
		border-radius: 50%;
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%) rotate(-90deg);
		width: var(--factor-width);
		height: var(--factor-width);
	}
	.sector-wrapper {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
	}
	.sector {
		/* sector-size (in percent) */
		--p: 20;
		/* offset from center */
		--b: 200px;
		--c: darkred;
		--o: 0deg;

		width: var(--sector-width);
		aspect-ratio: 1;
		position: relative;
		place-content: center;
		font-weight: bold;
		font-family: sans-serif;
		transform: rotate(var(--o));
		animation: p 1s 0.5s both;
	}
	.sector:before,
	.sector:after {
		content: '';
		position: absolute;
		border-radius: 50%;
	}
	.sector:before {
		inset: 0;
		background: radial-gradient(farthest-side, var(--c) 98%, #0000) top/var(--b) var(--b) no-repeat,
			conic-gradient(var(--c) calc(var(--p) * 1%), #0000 0);
		-webkit-mask: radial-gradient(
			farthest-side,
			#0000 calc(99% - var(--b)),
			#000 calc(100% - var(--b))
		);
		mask: radial-gradient(farthest-side, #0000 calc(100% - var(--b)), #000 calc(100% - var(--b)));
		border: 2px solid white;
	}
	.sector:after {
		inset: calc(50% - var(--b) / 2);
		background: var(--c);
		transform: rotate(calc(var(--p) * 3.6deg)) translateY(calc(50% - var(--w) / 2));
		border: 2px solid white;
	}
	.no-round:before {
		background-size: 0 0, auto;
	}
	.no-round:after {
		content: none;
	}
	.sector[data-factor='1'] {
		--c: var(--color-grey-400);
	}
	.sector[data-factor='2'] {
		--c: var(--color-orange-700);
	}
	.sector[data-factor='3'] {
		--c: var(--color-green-700);
	}
	.sector[data-factor='5'] {
		--c: var(--color-blue-700);
	}
	.sector[data-factor='7'] {
		--c: var(--color-purple-700);
	}
	.sector[data-factor='11'] {
		--c: var(--color-red-700);
	}

	.cellValue {
		background-color: var(--ctb);
		padding: var(--inner-padding);
		border-radius: var(--radius-full);
		transition-property: color, transform;
		transition-duration: 300ms;
		transition-timing-function: var(--easing-standard);
		position: relative;
		position: absolute;
		margin: auto;
		text-align: center;
	}
	.cellValue .inner {
		color: var(--color-black);
		--text-shadow-color: var(--color-grey-100);
		text-shadow: 
      /* outer contrast for readability */ -1px 1px 2px var(--text-shadow-color),
			/* outer contrast for readability */ 1px 1px 2px var(--text-shadow-color),
			/* outer contrast for readability */ 1px -1px 2px var(--text-shadow-color),
			/* outer contrast for readability */ -1px -1px 2px var(--text-shadow-color),
			/* Added shadow for depth */ 2px 4px 4px #282828;

		/* -webkit-text-stroke: 1px white; */
		/* text-shadow: 0px 4px 4px #282828; */
		position: absolute;
		margin: auto;
		text-align: center;
		font-size: max(min(var(--value-length), var(--max-font-size)), var(--min-font-size));
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		display: flex;
		justify-content: center;
		align-items: center;
		/* width: var(--factor-width); */
		/* max-width: var(--factor-width); */
		transform: translateY(-5%);
	}

	.cell.selected {
		background-color: var(--color-green-300);
		transform: scale(0.9);
	}
	.cell.selected .cellValue {
		transform: scale(1.2);
	}

	.cell.selectedLast:not(.evaluatesTo) {
		background-color: var(--color-green-500);
	}
	.cell.evaluatesTo.selected {
		background-color: var(--color-green-700);
	}
	.cell.no-eval:not(.evaluatesTo) {
		animation: shake 0.82s cubic-bezier(0.36, 0.07, 0.19, 0.97) both, grow-to-normal 0.82s linear,
			sepia 0.82s cubic-bezier(0.36, 0.07, 0.19, 0.97) both;
		transform: translate3d(0, 0, 0);
		backface-visibility: hidden;
		perspective: 1000px;
	}
	@keyframes grow-to-normal {
		0% {
			scale: 0.9;
		}
		80% {
			scale: 0.9;
		}
		1000% {
			scale: 0.9;
		}
	}
	@keyframes sepia {
		0% {
			filter: sepia(1);
		}
		80% {
			filter: sepia(1);
		}
		1000% {
		}
	}
	@keyframes shake {
		10%,
		90% {
			transform: translate3d(-1px, 0, 100px);
		}

		20%,
		80% {
			transform: translate3d(2px, 0, 0);
		}

		30%,
		50%,
		70% {
			transform: translate3d(-4px, 0, 0);
		}

		40%,
		60% {
			transform: translate3d(4px, 0, 0);
		}
	}
</style>
