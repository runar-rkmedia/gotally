<script lang="ts">
	export let open: boolean
	let el: HTMLDialogElement
	$: {
		if (el) {
			if (open) {
				if (!('showModal' in el)) {
					console.error('element does not support method "showModal"', el)
				} else {
					el.showModal()
				}
			} else {
				if (!('close' in el)) {
					console.error('element does not support method "close"', el)
				} else {
					el.close()
				}
			}
		}
	}
</script>

<dialog
	class:isOpen={open}
	bind:this={el}
	on:click={(e) => {
		if (e.target !== el) {
			return
		}
		console.log(e.currentTarget, e.target)
		open = false
	}}
>
	<div class="wrapper">
		<slot {open} />
	</div>
</dialog>

<style>
	dialog {
		width: 100%;
		background-color: transparent;
		border: none;
		margin: auto;
		opacity: 0;
		transition: opacity 0.5s var(--easing-standard);
	}
	.wrapper {
		height: min-content;
		width: 100%;
		background-color: #000038dd;
		color: var(--color-grey-50);
		box-shadow: var(--elevation-4);
		border-radius: var(--radius-md);
		backdrop-filter: blur(4px);
		overflow: auto;
		max-height: calc(100vh - env(safe-area-inset-bottom, 50px) - env(safe-area-inset-bottom, 50px));
	}
	@media (max-width: 640px) {
		dialog {
			min-width: 100%;
			min-height: 100%;
			padding: 0;
			margin: 0;
		}
		.wrapper {
			padding-bottom: 60px;
			padding-top: 60px;
			min-height: 100%;
			border-radius: unset;
		}
	}
	dialog.isOpen {
		opacity: 1;
		pointer-events: inherit;
	}
	dialog.isOpen::backdrop {
		opacity: 1;
	}
	dialog::backdrop {
		opacity: 0;
		transition: opacity 0.5s var(--easing-standard);
		background-color: rgba(0, 0, 255, 0.2);
		backdrop-filter: blur(2px);
	}
</style>
