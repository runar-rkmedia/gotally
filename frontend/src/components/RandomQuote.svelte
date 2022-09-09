<script lang="ts">
	let mode = 'random'
	let keyword = 'win'
	const key = import.meta.env.VITE_ZENQUOTES_API_KEY
	export let seed = Math.floor(Math.random() * 1e12)
	const quotes = [
		{
			paragraphs: ['A quitter never wins and a winner never quits.'],
			by: 'Napoleoon Hill'
		},
		{
			paragraphs: [
				'You’re not obligated to win.',
				'You’re obligated to keep trying.',
				'To the best you can do everyday.'
			],
			by: 'Jason Mraz'
		}
	]
	$: quote = quotes[seed % quotes.length]
</script>

<strong>Remember:</strong>
<figure>
	<blockquote>
		{#each quote.paragraphs as item}
			<!-- content here -->
			<p>
				{item}
			</p>
		{/each}
	</blockquote>
	<figcaption>&mdash {quote.by}</figcaption>
</figure>

<style>
	p {
		line-height: 1.4;
		margin-block: 0px;
	}
	blockquote {
		font-family: cursive;
		margin-inline: 0;
	}
	figure {
		text-align: end;
	}
</style>
