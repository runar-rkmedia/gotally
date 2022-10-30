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
		background-color: transparent;
		border: none;
		margin: auto;
		overflow: hidden;
		opacity: 0;
		transition: opacity 0.5s var(--easing-standard);
	}
	.wrapper {
		height: min-content;
		width: min-content;
		background-color: #00000088;
		color: var(--color-grey-50);
		box-shadow: var(--elevation-4);
		border-radius: var(--radius-md);
		backdrop-filter: blur(4px);
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
