<script lang="ts">
	import Field from './Field.svelte'
	import { storeHandler } from '../../connect-web/store'
	import BoardPreview from '../../components/BoardPreview.svelte'
	import type { ConnectError } from '@bufbuild/connect-web'
	import { GeneratorAlgorithm } from '../../connect-web'
	import { browser } from '$app/environment'
	import { onDestroy, onMount } from 'svelte'
	import { cellValue } from '../../components/board/cell'
	import { coordToIndex, indexToCoord } from '../../logic'
	import merge from 'lodash.merge'
	import GeneratedGameStats from './GeneratedGameStats.svelte'

	let result: Awaited<ReturnType<typeof storeHandler.generateGame>>[0]
	let resErr: ConnectError | Error | null
	let requestStart: Date | null = null
	let requestEnd: Date | null = null
	let timerMs = 0
	let loading = false
	enum tabs {
		generator = 1,
		preview,
		tips,
	}
	let state: {
		gen: Parameters<typeof storeHandler.generateGame>[0]
		tab: tabs
		autoChangeTab: boolean
	} = {
		tab: tabs.generator,
		autoChangeTab: false,
		gen: {
			rows: 3,
			columns: 3,
			targetCellValue: 64,
			maxBricks: 9,
			minMoves: 3,
			maxMoves: 9,
			withSolutions: true,
			randomCellChance: 0,
			algorithm: GeneratorAlgorithm.RANDOMIZED,
		},
	}
	const setCells = (rows: number, columns: number) =>
		new Array(rows * columns).fill(null).map((_, i) => ({ base: i, twopow: 0 }))
	const setAsTemplate: Parameters<typeof storeHandler.createTemplate>[0] = {
		name: 'New Challenge',
		description: '',
		targetCellValue: 64,
		idealMoves: 0,
		cells: setCells(3, 3),
		rows: 3,
		columns: 3,
		idealScore: 0,
		challengeNumber: undefined,
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
		const json = JSON.parse(v) as typeof state
		merge(state, json)
		state = state
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
		rows:
			!state.gen.rows ||
			state.gen.rows < 3 ||
			(state.gen.rows > 9 && 'Rows must be between 3 and 9'),
		columns:
			!state.gen.columns ||
			state.gen.columns < 3 ||
			(state.gen.columns > 9 && 'Columns must be between 3 and 9'),
		targetCellValue:
			!state.gen.targetCellValue ||
			state.gen.targetCellValue < 3 ||
			(state.gen.targetCellValue > 100_000_000 &&
				'TargetCellValue must be between 3 and 100 million'),
	} as Partial<Record<keyof typeof state.gen, boolean | string>>
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
			(!setAsTemplate.targetCellValue ||
				setAsTemplate.targetCellValue < 3 ||
				setAsTemplate.targetCellValue > 100_000_000) &&
			'TargetCellValue must be between 3 and 100 million',
	} as Partial<Record<keyof typeof setAsTemplate, boolean | string>>
	$: hasError = Object.values(error).some(Boolean)

	function saveState() {
		localStorage.setItem('gotally_generator', JSON.stringify(state))
	}

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
		const r = await storeHandler.commit(storeHandler.generateGame(state.gen))
		loading = false
		requestEnd = new Date()
		timerMs = requestStart.getTime() - requestEnd.getTime()
		console.log({ result: r, timerMs })
		result = r.result
		resErr = r.error
		if (!r.error && browser) {
			saveState()
		}
		if (result) {
			if (state.autoChangeTab) {
				state.tab = tabs.preview
			}
			setAsTemplate.idealMoves = Infinity
			for (const s of result.solutions) {
				if (s.moves < setAsTemplate.idealMoves) {
					setAsTemplate.idealMoves = s.moves
				}
			}
			setAsTemplate.targetCellValue = state.gen.targetCellValue
			setAsTemplate.description = `Get a target cell to a value of ${setAsTemplate.targetCellValue}. The game can be solved in ${setAsTemplate.idealMoves} moves`
			setAsTemplate.rows = state.gen.rows
			setAsTemplate.columns = state.gen.columns
			setAsTemplate.cells = result.game.board.cells
		}
	}
	const onSetAsTemplate = async () => {
		const r = await storeHandler.commit(storeHandler.createTemplate(setAsTemplate))
		resErr = r.error
	}
	const newGameFromTemplate = async () => {
		if (!setAsTemplate.rows) {
			return
		}
		const r = await storeHandler.commit(
			storeHandler.newGameFromTemplate({
				rows: setAsTemplate.rows,
				columns: setAsTemplate.columns,

				idealMoves: 0,
				idealScore: 0,

				targetCellValue: setAsTemplate.targetCellValue,
				name: setAsTemplate.name,
				description: setAsTemplate.description,
				cells: setAsTemplate.cells?.map((c) => ({
					base: c.base,
					twopow: c.twopow,
				})),
			})
		)
		if (r.error) {
			return
		}
		window.open('/', '_game_tab')
	}
	let previousColumns = 0
	$: {
		// When the board-size changes, we want to perform a resize
		// Coming from React, I am amazed that this is possible
		if (setAsTemplate.columns && setAsTemplate.rows && setAsTemplate.cells) {
			setAsTemplate.cells = resizeCells(
				previousColumns,
				setAsTemplate.columns,
				setAsTemplate.rows,
				setAsTemplate.cells.map((c) => ({ base: c.base || 0, twopow: c.twopow || 0 }))
			)
		}
		previousColumns = setAsTemplate.columns || 0
	}
	const resizeCells = (
		prevColumns: number,
		columns: number,
		rows: number,
		cells: { base: number; twopow: number }[]
	) => {
		const wantedCount = columns * rows
		const currentCount = cells.length

		if (wantedCount === currentCount) {
			return cells
		}
		const coordMap: Record<number, number | null> = {}
		for (let i = 0; i < currentCount; i++) {
			i
			const [x, y] = indexToCoord(i, prevColumns)
			const newIndex = coordToIndex(x, y, columns, rows)
			if (!newIndex) {
				continue
			}
			coordMap[newIndex] = i
		}
		return new Array(wantedCount).fill(null).map((_, i) => {
			const oldIndex = coordMap[i]
			if (oldIndex === null || oldIndex === undefined) {
				return { base: 0, twopow: 0 }
			}
			const cell = cells[oldIndex]
			if (cell) {
				return cell
			}
			return { base: 0, twopow: 0 }
		})
	}
</script>

<h1>Game generator</h1>

<div class="tabs">
	<button class:active={state.tab === tabs.generator} on:click={() => (state.tab = tabs.generator)}
		>1. Generator</button
	>
	<button class:active={state.tab === tabs.preview} on:click={() => (state.tab = tabs.preview)}
		>2. Preview</button
	>
	<button class:active={state.tab === tabs.tips} on:click={() => (state.tab = tabs.tips)}
		>Tips</button
	>
</div>

{#if resErr}
	<div>{resErr.message}</div>
	<!-- content here -->
{/if}
{#if state.tab === tabs.generator}
	<form on:submit|preventDefault={onSubmitForGeneration}>
		<div class="set">
			<label>
				Randomized - Slow, but gives more varied results
				<input
					name="algorithm"
					type="radio"
					bind:group={state.gen.algorithm}
					value={GeneratorAlgorithm.RANDOMIZED}
				/>
			</label>
			<label>
				Reverse - Fast, but very monotomous results
				<input
					name="algorithm"
					type="radio"
					bind:group={state.gen.algorithm}
					value={GeneratorAlgorithm.REVERSE}
				/>
			</label>
		</div>
		<div class="set">
			<Field error={error.rows} label="Rows">
				<input min="3" max="9" type="number" bind:value={state.gen.rows} />
			</Field>
			<Field error={error.columns} label="Column">
				<input min="3" max="9" type="number" bind:value={state.gen.columns} />
			</Field>
		</div>
		<Field error={error.targetCellValue} label="Target Cell Value">
			<input min="3" max="100000000" type="number" bind:value={state.gen.targetCellValue} />
		</Field>
		<Field error={error.maxAdditionalCells} label="Max Additional cells">
			<input min="-1" max="100000000" type="number" bind:value={state.gen.maxAdditionalCells} />
		</Field>
		<Field error={error.maxBricks} label="Max Bricks">
			<input min="3" max="100000000" type="number" bind:value={state.gen.maxBricks} />
		</Field>
		{#if state.gen.algorithm == GeneratorAlgorithm.REVERSE}
			<Field error={error.randomCellChance} label="Random cell chance">
				<input min="-1" max="120" type="number" bind:value={state.gen.randomCellChance} />
			</Field>
		{/if}
		<div class="set">
			<Field error={error.minMoves} label="Min moves">
				<input min="3" max="100000000" type="number" bind:value={state.gen.minMoves} />
			</Field>
			<Field error={error.maxMoves} label="Max moves">
				<input
					min={state.gen.minMoves}
					max="100000000"
					type="number"
					bind:value={state.gen.maxMoves}
				/>
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
	{#if result?.stats}
		<GeneratedGameStats stats={result.stats} />
	{/if}
{/if}
{#if state.tab === tabs.preview}
	<form on:submit|preventDefault={onSetAsTemplate}>
		<Field error={setAsTemplateError.name} label="Name">
			<input type="string" minlength="3" bind:value={setAsTemplate.name} />
		</Field>
		<Field error={setAsTemplateError.description} label="description">
			<input type="string" bind:value={setAsTemplate.description} />
		</Field>
		<div class="set">
			<Field error={setAsTemplateError.rows} label="Rows">
				<input min="3" max="9" type="number" bind:value={setAsTemplate.rows} />
			</Field>
			<Field error={setAsTemplateError.columns} label="Columns">
				<input min="3" max="9" type="number" bind:value={setAsTemplate.columns} />
			</Field>
		</div>
		<Field error={setAsTemplateError.challengeNumber} label="Challenge number">
			<input min="0" max="100000000" type="number" bind:value={setAsTemplate.challengeNumber} />
		</Field>
		<Field error={setAsTemplateError.targetCellValue} label="targetCellValue">
			<input type="string" bind:value={setAsTemplate.targetCellValue} />
		</Field>

		<button type="submit">Set as template</button>
		<button
			on:click|preventDefault={newGameFromTemplate}
			disabled={Object.values(setAsTemplateError).some(Boolean)}>Play (in tab)</button
		>
	</form>
{/if}
{#if state.tab === tabs.tips}
	<div class="tips">
		<p>
			Try to create challenges that create a unique <i>kind</i> of solution, not just changing the numbers/order.
			For instance, challenges vary by:
		</p>
		<ul>
			<li>
				The number of possible solutions, especially if the shortest solution is not too obvious.
			</li>
			<li>
				The <code>variations</code> within a solution is high. By <code>variations</code>, it is
				counted as the number of <i>different</i> actions are needed to solve the game, like
				swiping, multiplying, adding, or serial use of bricks (<code>twopow</code>).
			</li>
		</ul>
		<p>Use the generator with randomization for ideas.</p>
	</div>
{/if}

<div class="tmpFlexy">
	<details>
		<summary>Details</summary>
		<pre>{JSON.stringify({ options: state, setAsTemplate }, null, 2)}</pre>
		<pre>{JSON.stringify({ error }, null, 2)}</pre>
	</details>
	<div class="set">
		<label>
			Auto change tabs
			<input type="checkbox" bind:checked={state.autoChangeTab} on:change={saveState} />
		</label>
	</div>
	<button style="color: var(--color-white)" on:click={saveState}>Save state</button>
</div>
{#if state.tab === tabs.preview && setAsTemplate.cells && setAsTemplate.rows && setAsTemplate.columns}
	<div class="games">
		<div class="game">
			<BoardPreview
				on:cellclick={(e) => {
					console.log('cellclick', e.detail)
					if (!setAsTemplate.cells.length) {
						return
					}
					const n = Number(prompt('Change this cell', String(cellValue(e.detail.cell))))
					if (isNaN(n)) {
						console.log('foobar n', n)
						alert('must be a number')
						return
					}
					setAsTemplate.cells[e.detail.i] = { base: n, twopow: 0 }
				}}
				cells={setAsTemplate.cells.map((c) => ({ base: c.base || 0, twopow: c.twopow || 0 })) || []}
				rows={setAsTemplate.rows}
				columns={setAsTemplate.columns}
			/>
			<p>Cells: {setAsTemplate.cells.length}</p>
			<p>Columns: {setAsTemplate.columns}</p>
			<p>Rows: {setAsTemplate.rows}</p>
		</div>
	</div>
{/if}

<style>
	.tmpFlexy {
		display: flex;
		flex-direction: column;
		margin-top: 50px;
		overflow: scroll;
		max-height: 400px;
		position: fixed;
		top: 0;
		right: 0;
		background: var(--color-black);
		z-index: 1;
		color: var(--color-primary);
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
	button:disabled {
		background: var(--color-grey);
		opacity: 0.4;
	}
	/* Terribly looking tabs, but its just for internal use */
	.tabs {
		display: flex;
		position: relative;
		width: min-content;
	}
	.tabs::before {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 4px;
		z-index: -1;
		background: var(--color-primary);
	}
	.tabs button {
		background: var(--color-grey);
		transition: all 300ms var(--easing-standard);
		border-top-right-radius: 20px;
		border-top-left-radius: 20px;
		white-space: nowrap;
	}
	.tabs button.active {
		transform: translateY(-4px);
		background: var(--color-primary);
	}
	.tips {
		padding: var(--size-4);
	}
</style>
