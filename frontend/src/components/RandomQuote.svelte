<script lang="ts">
	let mode = 'random'
	let keyword = 'win'
	const key = import.meta.env.VITE_ZENQUOTES_API_KEY
	export let seed = Math.floor(Math.random() * 1e12)
	const quotes = [
		{
			paragraphs: ['A quitter never wins and a winner never quits.'],
			by: 'Napoleoon Hill',
		},
		{
			paragraphs: [
				'You’re not obligated to win.',
				'You’re obligated to keep trying.',
				'To the best you can do everyday.',
			],
			by: 'Jason Mraz',
		},
		{
			paragraphs: [
				'The brain is like a muscle.',
				'When it is in use we feel very good.',
				'Understanding is joyous.',
			],
			by: 'Carl Sagan',
		},
		{
			paragraphs: ['You are doing great,', 'keep those brainfluids flowing!'],
		},
	]
	$: quote = quotes[seed % quotes.length]
</script>

<figure>
	<blockquote>
		{#each quote.paragraphs as item}
			<!-- content here -->
			<p>
				{item}
			</p>
		{/each}
	</blockquote>
	{#if quote.by}
		<figcaption>&mdash {quote.by}</figcaption>
	{/if}
</figure>

<style>
	p {
		line-height: 1.4;
		margin-block: 0px;
	}
	blockquote {
		font-size: 1.2rem;
		font-family: cursive;
		margin-inline: 0;
	}
</style>
