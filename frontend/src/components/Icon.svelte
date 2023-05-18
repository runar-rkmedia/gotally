<script lang="ts">
	// import Cog from 'lucide-svelte/dist/svelte/icons/cog.svelte'
	import {
		Cog,
		GithubIcon,
		GraduationCapIcon,
		HelpCircle,
		InfinityIcon,
		LockIcon,
		PlayIcon,
		RefreshCcw,
		StarHalfIcon,
		StarIcon,
		Sword,
		Undo2,
	} from 'lucide-svelte'
	export let icon:
		| 'settings'
		| 'undo'
		| 'help'
		| 'tutorial'
		| 'challenge'
		| 'infinite'
		| 'restart'
		| 'continue'
		| 'github'
		| 'play'
		| 'star'
		| 'star-half'
		| 'lock'
	export let color: Color | undefined = undefined
	type Color = 'primary' | 'secondary' | 'danger' | 'success' | 'error' | 'white' | 'black'
	// your script goes here

	const getColor = (color: Color) => {
		switch (color) {
			case 'primary':
			case 'secondary':
			case 'danger':
			case 'success':
			case 'error':
			case 'white':
			case 'black':
				return `var(--color-${color})`
		}
	}
	$: c = (color && getColor(color)) || 'currentColor'
	let size = '1.8em'
	// https://lucide.dev/?search=
</script>

{#if icon === 'settings'}
	<Cog color={c} {size} />
{:else if icon === 'help'}
	<HelpCircle color={c} {size} />
{:else if icon === 'undo'}
	<Undo2 color={c} {size} />
{:else if icon === 'tutorial'}
	<GraduationCapIcon color={c} {size} />
{:else if icon === 'challenge'}
	<Sword color={c} {size} />
{:else if icon === 'infinite'}
	<InfinityIcon color={c} {size} />
{:else if icon === 'restart'}
	<RefreshCcw color={c} {size} />
{:else if icon === 'continue'}
	<RefreshCcw color={c} {size} />
{:else if icon === 'github'}
	<GithubIcon color={c} {size} />
{:else if icon === 'play'}
	<PlayIcon color={c} {size} />
{:else if icon === 'star'}
	<StarIcon color={c} {size} />
{:else if icon === 'star-half'}
	<StarHalfIcon color={c} {size} />
{:else if icon === 'lock'}
	<LockIcon color={c} {size} />
{:else}
	No icon named {icon}
{/if}

<style>
	:global(svg) {
		display: inline-block;
		vertical-align: bottom;
	}
</style>
