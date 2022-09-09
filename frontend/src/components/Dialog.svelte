<script lang="ts">
	export let open: boolean
</script>

<dialog {open}>
	<slot />
</dialog>

<style>
	dialog {
		z-index: 1;
		pointer-events: none;
		opacity: 0;
		transition: opacity 0.5s;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		background-color: var(--color-grey-800);
		color: var(--color-grey-50);
		border: none;
	}
	dialog[open] {
		opacity: 1;
		pointer-events: inherit;
	}
	dialog::backdrop {
		background-color: rgba(0, 0, 255, 0.2);
	}
</style>
