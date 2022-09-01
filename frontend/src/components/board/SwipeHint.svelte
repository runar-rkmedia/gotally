<script lang="ts">
	import { SwipeDirection } from '../../connect-web'

	export let instruction: SwipeDirection
	export let active: boolean
</script>

<div class="wrapper" class:active>
	<div
		class="shadow"
		class:up={instruction === SwipeDirection.UP}
		class:right={instruction === SwipeDirection.RIGHT}
		class:down={instruction === SwipeDirection.DOWN}
		class:left={instruction === SwipeDirection.LEFT}
	>
		<div class="triangle-wrapper">
			<div class="triangle" />
		</div>
	</div>
</div>

<style>
	/* Glow-effect by Dave Brogan https://codepen.io/davebrogan/pen/OJJKXpy */
	.wrapper {
		--vw-width: 30vw;
		--vw-height: calc(var(--vw-width) / (1 / 1));
		--tod-tri: #77fc47;
		pointer-events: none;
		z-index: 1;
		/* opacity: 0.9; */
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		top: 0;
		opacity: 0;
		display: flex;
		justify-content: center;
		align-items: center;
		transform: scale(0.4);
	}
	.active {
		opacity: 1;
	}
	@keyframes bounce {
		0%,
		20%,
		55%,
		75%,
		100% {
			-webkit-transform: translateY(0);
		}
		40% {
			-webkit-transform: translateY(-125px);
		}
		60% {
			-webkit-transform: translateY(-85px);
		}
	}
	@keyframes pulse {
		0% {
			transform: scale(1);
		}
		50% {
			transform: scale(1.1);
		}
		100% {
			transform: scale(1);
		}
	}

	.triangle-wrapper {
		animation-name: pulse;
		animation-fill-mode: both;
		animation-duration: 1.3s;
		animation-iteration-count: infinite;
		width: var(--vw-width);
		height: var(--vw-height);
		clip-path: polygon(50% 0%, 0% 100%, 100% 100%);
		background-color: var(--tod-tri);
		display: flex;
		justify-content: center;
		align-items: center;
		animation-name: bounce;
		animation-fill-mode: both;
		animation-duration: 1.3s;
	}

	.left {
		transform: rotate(-90deg);
	}

	.right {
		transform: rotate(90deg);
	}
	.down {
		transform: rotate(180deg);
	}
	.shadow {
		filter: drop-shadow(10px 10px 200px var(--tod-tri)) drop-shadow(-10px -10px 50px var(--tod-tri));
	}
	.triangle {
		width: calc(var(--vw-width) - 30px);
		height: calc(var(--vw-width) - 30px);
		background-color: hsla(210, 50%, 14%);
		clip-path: polygon(50% 0%, 0% 100%, 100% 100%);
		filter: blur(115px) drop-shadow(-10px -10px 75px var(--tod-tri));
	}
</style>
