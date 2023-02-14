<script lang="ts">
	export let error: string | boolean | undefined
	export let label: string | boolean
	// your script goes here
</script>

<div>
	<label
		><slot name="label">{label}:</slot>
		<slot />
		{#if error}
			<div class="error">{error}</div>
		{/if}
	</label>
</div>

<style>
	label {
		display: flex;
		flex-direction: column;
	}
	.error {
		color: var(--color-error);
	}
</style>
