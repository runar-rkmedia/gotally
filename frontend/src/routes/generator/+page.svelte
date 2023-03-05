<script lang="ts">
	import Field from './Field.svelte'
	import { storeHandler } from '../../connect-web/store'
	import BoardPreview from '../../components/BoardPreview.svelte'
	import type { ConnectError } from '@bufbuild/connect-web'
	import { GeneratorAlgorithm } from '../../connect-web'
	import { browser } from '$app/environment'
	import { onDestroy, onMount } from 'svelte'

	let result: Awaited<ReturnType<typeof storeHandler.generateGame>>[0]
	let resErr: ConnectError | Error | null
	let requestStart: Date | null = null
	let requestEnd: Date | null = null
	let timerMs = 0
	let loading = false
	const gen: Parameters<typeof storeHandler.generateGame>[0] = {
		rows: 3,
		columns: 3,
		targetCellValue: 64,
		maxBricks: 9,
		minMoves: 3,
		maxMoves: 9,
		withSolutions: true,
		randomCellChance: 0,
		algorithm: GeneratorAlgorithm.RANDOMIZED
	}
	const setAsTemplate: Parameters<typeof storeHandler.createTemplate>[0] = {
		name: 'New Challenge',
		description: '',
		targetCellValue: 0,
		idealMoves: 0,
		cells: undefined,
		rows: 0,
		columns: 0
	}
	onMount(() => {
		if (!browser) {
			return
		}
		// should probably use a store instead here, but this is only for development
		const v = localStorage.getItem('gotally_generator')
		if (!v) {
			return
		}
		const json = JSON.parse(v) as typeof gen
		for (const [k, v] of Object.entries(json)) {
			;(gen as any)[k] = v
		}
	})

	const interval = setInterval(() => {
		if (requestEnd) {
			return
		}
		if (!requestStart) {
			timerMs = 0
			return
		}
		timerMs = new Date().getTime() - requestStart.getTime()
	}, 500)

	onDestroy(() => {
		clearInterval(interval)
	})
	$: error = {
		rows: !gen.rows || gen.rows < 3 || (gen.rows > 9 && 'Rows must be between 3 and 9'),
		columns:
			!gen.columns || gen.columns < 3 || (gen.columns > 9 && 'Columns must be between 3 and 9'),
		targetCellValue:
			!gen.targetCellValue ||
			gen.targetCellValue < 3 ||
			(gen.targetCellValue > 100_000_000 && 'TargetCellValue must be between 3 and 100 million')
	} as Partial<Record<keyof typeof gen, boolean | string>>
	$: setAsTemplateError = {
		rows:
			!setAsTemplate.rows ||
			setAsTemplate.rows < 3 ||
			(setAsTemplate.rows > 9 && 'Rows must be between 3 and 9'),
		columns:
			!setAsTemplate.columns ||
			setAsTemplate.columns < 3 ||
			(setAsTemplate.columns > 9 && 'Columns must be between 3 and 9'),
		targetCellValue:
			!setAsTemplate.targetCellValue ||
			setAsTemplate.targetCellValue < 3 ||
			(setAsTemplate.targetCellValue > 100_000_000 &&
				'TargetCellValue must be between 3 and 100 million')
	} as Partial<Record<keyof typeof setAsTemplate, boolean | string>>
	$: hasError = Object.values(error).some(Boolean)

	const onSubmitForGeneration = async () => {
		if (hasError || loading) {
			return
		}
		result = null
		resErr = null
		timerMs = 0
		requestStart = new Date()
		requestEnd = null
		loading = true
		const r = await storeHandler.commit(storeHandler.generateGame(gen))
		loading = false
		requestEnd = new Date()
		timerMs = requestStart.getTime() - requestEnd.getTime()
		console.log({ result: r, timerMs })
		result = r.result
		resErr = r.error
		if (!r.error && browser) {
			localStorage.setItem('gotally_generator', JSON.stringify(gen))
		}
		if (result) {
			setAsTemplate.idealMoves = Infinity
			for (const s of result.solutions) {
				if (s.moves < setAsTemplate.idealMoves) {
					setAsTemplate.idealMoves = s.moves
				}
			}
			setAsTemplate.targetCellValue = gen.targetCellValue
			setAsTemplate.description = `Get a target cell to a value of ${setAsTemplate.targetCellValue}. The game can be solved in ${setAsTemplate.idealMoves} moves`
			setAsTemplate.rows = gen.rows
			setAsTemplate.columns = gen.columns
			setAsTemplate.cells = result.game.board.cells
		}
	}
	const onSetAsTemplate = async () => {
		const r = await storeHandler.commit(storeHandler.createTemplate(setAsTemplate))
		resErr = r.error
	}
</script>

<h1>Game generator</h1>

{#if resErr}
	<div>{resErr.message}</div>
	<!-- content here -->
{/if}
<form on:submit|preventDefault={onSubmitForGeneration}>
	<Field error={error.algorithm} label="algorithm">
		<select name="algorithm" bind:value={gen.algorithm}>
			<option value={GeneratorAlgorithm.RANDOMIZED}
				>Randomized - Slow, but gives more varied results</option
			>
			<option value={GeneratorAlgorithm.REVERSE}>Reverse - Fast, but very monotomous results</option
			>
		</select>
	</Field>
	<div class="set">
		<Field error={error.rows} label="Rows">
			<input min="3" max="80" type="number" bind:value={gen.rows} />
		</Field>
		<Field error={error.columns} label="Column">
			<input min="3" max="80" type="number" bind:value={gen.columns} />
		</Field>
	</div>
	<Field error={error.targetCellValue} label="Target Cell Value">
		<input min="3" max="100000000" type="number" bind:value={gen.targetCellValue} />
	</Field>
	<Field error={error.maxAdditionalCells} label="Max Additional cells">
		<input min="-1" max="100000000" type="number" bind:value={gen.maxAdditionalCells} />
	</Field>
	<Field error={error.maxBricks} label="Max Bricks">
		<input min="3" max="100000000" type="number" bind:value={gen.maxBricks} />
	</Field>
	{#if gen.algorithm == GeneratorAlgorithm.REVERSE}
		<Field error={error.randomCellChance} label="Random cell chance">
			<input min="-1" max="120" type="number" bind:value={gen.randomCellChance} />
		</Field>
	{/if}
	<div class="set">
		<Field error={error.minMoves} label="Min moves">
			<input min="3" max="100000000" type="number" bind:value={gen.minMoves} />
		</Field>
		<Field error={error.maxMoves} label="Max moves">
			<input min={gen.minMoves} max="100000000" type="number" bind:value={gen.maxMoves} />
		</Field>
	</div>
	<button type="submit" disabled={hasError || loading}>Send</button>
	{#if loading}
		{#if timerMs}
			<p>Waiting for game-generation {timerMs}ms</p>
		{/if}
	{:else if timerMs}
		<p>Game generated in {timerMs}ms</p>
	{/if}
</form>
{#if result}
	<form on:submit|preventDefault={onSetAsTemplate}>
		<Field error={setAsTemplateError.name} label="Name">
			<input type="string" minlength="3" bind:value={setAsTemplate.name} />
		</Field>
		<Field error={setAsTemplateError.description} label="description">
			<input type="string" bind:value={setAsTemplate.description} />
		</Field>
		<Field error={setAsTemplateError.challengeNumber} label="Challenge number">
			<input min="0" max="100000000" type="number" bind:value={setAsTemplate.challengeNumber} />
		</Field>

		<button type="submit">Set as template</button>
	</form>
{/if}

<div class="tmpFlexy">
	<pre>{JSON.stringify({ options: gen, setAsTemplate }, null, 2)}</pre>
	<pre>{JSON.stringify({ error }, null, 2)}</pre>
</div>
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
		grid-template-columns: 1fr 1fr;
	}
	.game {
		border: 1px solid hotpink;
	}
	form {
		max-width: 22ch;
	}
	button {
		background: var(--color-primary);
		padding: var(--size-2) var(--size-3);
	}
</style>
