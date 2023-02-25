<script lang="ts">
	import Field from './Field.svelte'
	import { storeHandler } from '../../connect-web/store'
	import BoardPreview from '../../components/BoardPreview.svelte'
	import type { ConnectError } from '@bufbuild/connect-web'

	let result: Awaited<ReturnType<typeof storeHandler.generateGame>>[0]
	let resErr: ConnectError | Error | null
	const o: Parameters<typeof storeHandler.generateGame>[0] = {
		rows: 3,
		columns: 3,
		targetCellValue: 64,
		maxBricks: 9,
		minMoves: 3,
		maxMoves: 9,
		withSolutions: true,
		randomCellChance: 0
	}
	$: error = {
		rows: !o.rows || o.rows < 3 || (o.rows > 9 && 'Rows must be between 3 and 9'),
		columns: !o.columns || o.columns < 3 || (o.columns > 9 && 'Columns must be between 3 and 9'),
		targetCellValue:
			!o.targetCellValue ||
			o.targetCellValue < 3 ||
			(o.targetCellValue > 100_000_000 && 'TargetCellValue must be between 3 and 100 million')
	} as Partial<Record<keyof typeof o, boolean | string>>
	$: hasError = Object.values(error).some(Boolean)

	const onSubmit = async () => {
		if (hasError) {
			return
		}
		result = null
		resErr = null
		const r = await storeHandler.commit(storeHandler.generateGame(o))
		console.log({ result: r })
		result = r.result
		resErr = r.error
	}
</script>

<h1>Game generator</h1>

{#if resErr}
	<div>{resErr.message}</div>
	<!-- content here -->
{/if}
<form on:submit|preventDefault={onSubmit}>
	<div class="set">
		<Field error={error.rows} label="Rows">
			<input min="3" max="80" type="number" bind:value={o.rows} />
		</Field>
		<Field error={error.columns} label="Column">
			<input min="3" max="80" type="number" bind:value={o.columns} />
		</Field>
	</div>
	<Field error={error.targetCellValue} label="Target Cell Value">
		<input min="3" max="100000000" type="number" bind:value={o.targetCellValue} />
	</Field>
	<Field error={error.maxAdditionalCells} label="Max Additional cells">
		<input min="-1" max="100000000" type="number" bind:value={o.maxAdditionalCells} />
	</Field>
	<Field error={error.maxBricks} label="Max Bricks">
		<input min="3" max="100000000" type="number" bind:value={o.maxBricks} />
	</Field>
	<Field error={error.randomCellChance} label="Random cell chance">
		<input min="-1" max="120" type="number" bind:value={o.randomCellChance} />
	</Field>
	<div class="set">
		<Field error={error.minMoves} label="Min moves">
			<input min="3" max="100000000" type="number" bind:value={o.minMoves} />
		</Field>
		<Field error={error.maxMoves} label="Max moves">
			<input min={o.minMoves} max="100000000" type="number" bind:value={o.maxMoves} />
		</Field>
	</div>
	<button type="submit" disabled={hasError}>Send</button>
	{#if result}
		<div class="games">
			<div class="game">
				<BoardPreview
					cells={result.game.board.cells}
					rows={result.game.board.rows}
					columns={result.game.board.columns}
				/>
			</div>
			{#if result.solutions?.length}
				{#each result.solutions as s}
					{#if s.board}
						<!-- content here -->
						<div class="game">
							<BoardPreview cells={s.board.cells} rows={s.board.rows} columns={s.board.columns} />
							Score: {s.score} -- Moves: {s.moves}
						</div>
					{/if}
				{/each}
			{/if}
		</div>
	{/if}
</form>

<div class="tmpFlexy">
	<pre>{JSON.stringify({ result }, null, 2)}</pre>
	<pre>{JSON.stringify({ options: o }, null, 2)}</pre>
	<pre>{JSON.stringify({ error }, null, 2)}</pre>
</div>

<style>
	.tmpFlexy {
		display: flex;
		margin-top: 50px;
		overflow: scroll;
		max-height: 400px;
	}
	.set {
		display: flex;
		gap: 10px;
	}
	.games {
		display: grid;
		gap: 10px;
	}
	.game {
		border: 1px solid hotpink;
	}
	form {
		max-width: 22ch;
	}
</style>
