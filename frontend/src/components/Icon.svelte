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
		if (!color) {
			return 'currentColor'
		}
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
	let Ic = Cog

	switch (icon) {
		case 'settings':
			Ic = Cog
			break
		case 'help':
			Ic = HelpCircle
			break
		case 'undo':
			Ic = Undo2
			break
		case 'tutorial':
			Ic = GraduationCapIcon
			break
		case 'challenge':
			Ic = Sword
			break
		case 'infinite':
			Ic = InfinityIcon
			break
		case 'restart':
			Ic = RefreshCcw
			break
		case 'continue':
			Ic = RefreshCcw
			break
		case 'github':
			Ic = GithubIcon
			break
		case 'play':
			Ic = PlayIcon
			break
		case 'star':
			Ic = StarIcon
			break
		case 'star-half':
			Ic = StarHalfIcon
			break
		case 'lock':
			Ic = LockIcon
			break
		default:
			throw new Error('unmapped icon ' + icon)
	}
</script>

<Ic
	color={c}
	{size}
	class={$$props.class === undefined ? 'mr-2 -ml-1 w-5 h-5' : $$props.class}
	{...$$props}
/>

<style>
	:global(svg) {
		display: inline-block;
		vertical-align: bottom;
	}
</style>
