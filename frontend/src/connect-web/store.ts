import type { ConnectError, CallOptions } from '@bufbuild/connect-web'
import type Buf from '@bufbuild/protobuf'
import { writable } from 'svelte/store'
import { client, go, handleError, type GoPromise } from './client'
import type { Instruction, SwipeDirection } from './proto/tally/v1/board_pb'
import type Api from './proto/tally/v1/board_pb'
import { objectKeys } from 'simplytyped'
import type { Message, PartialMessage } from '@bufbuild/protobuf'
// import type { Board } from './proto/tally/v1/board_pb'

interface Store {
	session: Session
	usersVotes: Record<string, Vote | undefined>
	didWin: boolean
	hints: Instruction[]
	hintDoneIndex: number
}

export type Vote = Strip<Api.VoteBoardResponse>
export type Cell = Replaced<Strip<Api.Cell>, bigint, number>
export type Challenge = Replaced<Strip<Api.GameChallenge>, bigint, number>
export type Board = Replaced<Omit<Strip<Api.Board>, 'cells'> & { cells: Cell[] }, bigint, number>
export type Game = Replaced<Omit<Strip<Api.Game>, 'board'> & { board: Board }, bigint, number>
export type Session = Replaced<Omit<Strip<Api.Session>, 'game'> & { game: Game }, bigint, number>
export type GeneratedGame = Replaced<
	Omit<Strip<Api.GenerateGameResponse>, 'game'> & { game: Game },
	bigint,
	number
>

type Strip<T extends Buf.Message> = Omit<
	T,
	| 'equals'
	| 'clone'
	| 'fromBinary'
	| 'fromJson'
	| 'fromJsonString'
	| 'toBinary'
	| 'toJson'
	| 'toJsonString'
	| 'getType'
>
type Primitive = string | number | bigint | boolean | null | undefined

type Replaced<T, TReplace, TWith, TKeep = Primitive> = T extends TReplace | TKeep
	? T extends TReplace
		? TWith | Exclude<T, TReplace>
		: T
	: {
			[P in keyof T]: Replaced<T[P], TReplace, TWith, TKeep>
	  }

const strip = <ReturnType>(obj: any): ReturnType => {
	const {
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		equals,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		clone,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		fromBinary,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		fromJson,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		fromJsonString,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		toBinary,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		toJson,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		toJsonString,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		getType,
		//eslint-enable
		...rest
	} = obj
	const deep = objectKeys(rest).reduce((r, k) => {
		const v = rest[k]
		switch (typeof v) {
			case 'function':
				return r
			case 'bigint':
				// console.log('stripping bigint', v, Number(v))
				return { ...r, [k]: Number(v) }
			case 'object':
				if (Array.isArray(v)) {
					return {
						...r,
						[k]: v.map((vv) => {
							switch (typeof vv) {
								case 'string':
								case 'number':
									return vv
								case 'bigint':
									return Number(vv)
							}
							return strip(vv)
						}),
					}
				}
				return { ...r, [k]: strip(v as any) }
			default:
				return { ...r, [k]: rest[k] }
		}
	}, {})

	return deep as any as ReturnType
}
export const store = writable<Store>({
	hints: [],
	didWin: false,
	hintDoneIndex: -1,
	session: null as any,
	usersVotes: {},
})
/** It is a bit annoying that the types include some extra methods, that I do not need
 * This strips these away
 * */
type OnlyMessage<T extends Message> = Omit<
	Replaced<PartialMessage<T>, bigint, number>,
	keyof Message
>
export interface ApiType {
	swipe: (
		direction: SwipeDirection
	) => CommitableGoResult<{ board: Board; moves: number; didChange: boolean }>
	generateGame: (
		payload: Replaced<OnlyMessage<Api.GenerateGameRequest>, bigint, number>
	) => CommitableGoResult<GeneratedGame>
	createTemplate: (
		payload: Omit<OnlyMessage<Api.CreateGameChallengeRequest>, 'cells'> & {
			cells: OnlyMessage<Api.CreateGameChallengeRequest['cells'][0]>[]
		}
	) => CommitableGoResult<{ id: string; challengeNumber: number }>
	getChallenges: (
		payload: Replaced<OnlyMessage<Api.GetGameChallengesRequest>, bigint, number>
	) => CommitableGoResult<Challenge[]>
	vote: (options: OnlyMessage<Api.VoteBoardRequest>) => CommitableGoResult<Vote>
	getSession: () => CommitableGoResult<Session>
	combineCells: (
		selection: number[]
	) => CommitableGoResult<{ didWin: boolean; board: Board; moves: number; score: number }>
	restartGame: () => CommitableGoResult<{ board: Board; moves: number; score: number }>
	undo: () => CommitableGoResult<{ board: Board; moves: number; score: number }>
	newGame: (
		options: OnlyMessage<Api.NewGameRequest>
	) => CommitableGoResult<{ board: Board; moves: number; score: number }>
	newGameFromTemplate: (
		options: OnlyMessage<Api.NewGameFromTemplateRequest>
	) => CommitableGoResult<{ board: Board; moves: number; score: number }>
	getHint: (
		options?: OnlyMessage<Api.GetHintRequest>
	) => CommitableGoResult<{ instructions: Api.Instruction[] }>
	/** utility-function used along with CommitableGoResult-promises to auto-commit when the caller wishes to commit the result when it is ready */
	commit: <T = any>(
		promise: CommitableGoResult<T>
	) => Promise<{ result: T | null; error: ConnectError | Error | null }>
}

// The request was valid, but resulted in no change
export const ErrNoChange = new Error('no change')
// Result of a failed null-check
export const ErrNoResult = new Error('no result')

type ApiMethodKeys = keyof Omit<Record<keyof ApiType, null>, 'commit'>

class ApiStore implements ApiType {
	commit: ApiType['commit'] = async (promise) => {
		const [result, commit, error] = await promise
		if (error) {
			return { result, error }
		}
		await commit()
		return { result, error }
	}
	// sigh, I really dislike this mapping and stripping...
	// there should be a better way to automate this...
	// For now, this type is not really important, it is only used internallly
	generateGame: ApiType['generateGame'] = async (options) => {
		const [result, err] = await go(client.generateGame(options as any))
		if (err) {
			handleError('generateGame', err)
			return [null, null, err]
		}
		const { game } = result
		if (!game) {
			return [null, null, ErrNoResult]
		}
		const res = {
			game: strip<Game>(game),
			idealMoves: Number(result.idealMoves),
			idealScore: Number(result.idealScore),
			highestScore: Number(result.highestScore),
			solutions: result.solutions.map((g) => strip<Game>(g)),
			stats: strip<any>(result.stats),
		}
		const commit = async () => {
			// TODO: commit the result to the store
			console.error('not implemented: commit for generateGame', { res })
		}

		return [res, commit, null]
	}
	getChallenges: ApiType['getChallenges'] = async (options) => {
		const [result, err] = await go(client.getGameChallenges(options as any))
		if (err) {
			handleError('getChallenges', err)
			return [null, null, err]
		}
		const res = result.challenges.map((r) => strip<Challenge>(r))
		const commit = async () => {
			// TODO: commit the result to the store
			console.error('not implemented: commit for getChallenges', { res })
		}

		return [res, commit, null]
	}
	createTemplate: ApiType['createTemplate'] = async (options) => {
		const [result, err] = await go(client.createGameChallenge(options as any))
		if (err) {
			handleError('createTemplate', err)
			return [null, null, err]
		}
		const res = {
			id: result.id,
			challengeNumber: result.challengeNumber,
		}
		const commit = async () => {
			// TODO: commit the result to the store
			console.error('not implemented: commit for createTemplate', { res })
		}

		return [res, commit, null]
	}
	getHint: ApiType['getHint'] = async (options = {}) => {
		const [result, err] = await go(client.getHint(options))
		if (err) {
			handleError('getHint', err)
			return [null, null, err]
		}
		const { instructions } = result
		if (!instructions) {
			return [null, null, ErrNoResult]
		}
		const res = {
			instructions,
		}
		const commit = async () => {
			store.update((s) => ({
				...s,
				hints: result.instructions,
				hintDoneIndex: -1,
			}))
		}
		return [res, commit, null]
	}
	newGame: ApiType['newGame'] = async (options) => {
		const s = await go(client.newGame(options))
		return this._newGame(s, options)
	}
	newGameFromTemplate: ApiType['newGameFromTemplate'] = async (options) => {
		const s = await go(client.newGameFromTemplate(options))
		return this._newGame(s, options)
	}
	/** Takes in the response from the api, validates for error and retuns as needed */
	_newGame = (
		s: Awaited<GoPromise<Api.NewGameResponse>>,
		options: any
	): Awaited<
		CommitableGoResult<{ board: Board; moves: number; score: number; description: string }>
	> => {
		const [result, err] = s

		console.debug('Creating new game', { options, result, err })
		if (err) {
			handleError('newGame', err)
			return [null, null, err]
		}
		const { board: _board } = result
		if (!_board) {
			return [null, null, ErrNoResult]
		}
		const board = strip<Board>(_board)
		const res = {
			board,
			moves: Number(result.moves),
			score: Number(result.score),
			description: result.description,
		}
		const commit = async () => {
			store.update((s) => {
				return {
					...s,
					didWin: false,
					hints: [],
					hintDoneIndex: -1,
					session: {
						...s.session,
						game: {
							...s.session?.game,
							...res,
						},
					},
				}
			})
		}
		return [res, commit, null]
	}
	restartGame: ApiType['restartGame'] = async () => {
		const [result, err] = await go(client.restartGame({}))
		if (err) {
			handleError('restartGame', err)
			return [null, null, err]
		}
		const { board: _board } = result
		if (!_board) {
			return [null, null, ErrNoResult]
		}
		const board = strip<Board>(_board)
		// response.board = result.board
		const res = {
			board,
			moves: Number(result.moves),
			score: Number(result.score),
		}
		const commit = async () => {
			store.update((s) => ({
				...s,
				didWin: false,
				hints: [],
				hintDoneIndex: -1,
				session: {
					...s.session,
					game: {
						...s.session.game,
						...res,
					},
				},
			}))
		}
		return [res, commit, null]
	}
	undo: ApiType['undo'] = async () => {
		const [result, err] = await go(client.undo({}))
		if (err) {
			handleError('undo', err)
			return [null, null, err]
		}
		const { board: _board } = result
		if (!_board) {
			return [null, null, ErrNoResult]
		}
		const board = strip<Board>(_board)
		// response.board = result.board
		const res = {
			board,
			moves: Number(result.moves),
			score: Number(result.score),
		}
		const commit = async () => {
			store.update((s) => ({
				...s,
				didWin: false,
				hints: [],
				hintDoneIndex: -1,
				session: {
					...s.session,
					game: {
						...s.session.game,
						...res,
					},
				},
			}))
		}
		return [res, commit, null]
	}
	combineCells: ApiType['combineCells'] = async (selection: number[]) => {
		const [result, err] = await go(
			client.combineCells({
				selection: {
					case: 'indexes',
					value: { index: selection },
				},
			})
		)
		if (err) {
			handleError('combineCells', err)
			return [null, null, err]
		}
		const { board: _board } = result
		if (!_board) {
			return [null, null, ErrNoResult]
		}
		const board = strip<Board>(_board)

		const res = {
			moves: Number(result.moves),
			score: Number(result.score),
			board: board,
			didWin: result.didWin,
		}
		const commit = async () => {
			store.update((s) => {
				const next: Store = {
					...s,
					hintDoneIndex: -1,
					didWin: result.didWin,
					hints: [],
					session: {
						...s.session,
						game: {
							...s.session.game,
							moves: res.moves,
							score: res.score,
							board: board,
						},
					},
				}
				const nextHint = s.hints[s.hintDoneIndex + 1]
				if (nextHint && nextHint.instructionOneof.case === 'combine') {
					const equal =
						nextHint && nextHint.instructionOneof.value.index.join() === selection.join()

					if (equal) {
						next.hintDoneIndex = s.hintDoneIndex + 1
						next.hints = s.hints
					}
				}
				return next
			})
		}
		return [res, commit, null]
	}
	getSession: ApiType['getSession'] = async () => {
		const sessionID = localStorage.getItem('sessionID') || ''
		const options: CallOptions = {
			onHeader: (h) => {
				const auth = h.get('Authorization')
				if (!auth) {
					return
				}
				localStorage.setItem('sessionID', auth)
			},
		}
		if (sessionID) {
			options.headers = {
				...options.headers,
				Authorization: sessionID,
			}
		}
		const [res, err] = await go(client.getSession({}, options))
		if (err) {
			handleError('getSession', err)
			return [null, null, err]
		}
		const { session: _session } = res
		if (!_session) {
			return [null, null, ErrNoResult]
		}
		const result = strip<Session>(_session)
		const commit = async () => {
			store.update((s) => ({ ...s, session: result as any }))
		}
		return [result, commit, err]
	}
	vote: ApiType['vote'] = async (options) => {
		const [result, err] = await go(client.voteBoard(options))
		if (err) {
			handleError('vote', err)
			return [null, null, err]
		}
		const vote = strip<Vote>(result)
		const commit = async () => {
			store.update((s) => ({
				...s,
				usersVotes: { ...s.usersVotes, [result.id]: vote },
			}))
		}
		return [vote, commit, null]
	}

	swipe: ApiType['swipe'] = async (direction) => {
		const [result, err] = await go(client.swipeBoard({ direction }))
		if (err) {
			handleError('swipe', err)
			return [null, null, err]
		}
		if (!result.didChange) {
			return [null, null, ErrNoChange]
		}
		const { board: _board } = result
		if (!_board) {
			return [null, null, ErrNoResult]
		}

		const board = strip<Board>(_board)

		const res = {
			board,
			moves: Number(result.moves),
			didChange: result.didChange,
		}

		const commit = async () => {
			store.update((s) => {
				const next = {
					...s,
					session: {
						...s.session,
						game: {
							...s.session.game,
							moves: res.moves,
							board: board,
						},
					},
				}
				const nextHint = s.hints[s.hintDoneIndex + 1]

				if (
					nextHint &&
					nextHint.instructionOneof.case === 'swipe' &&
					nextHint.instructionOneof.value === direction
				) {
					next.hintDoneIndex = next.hintDoneIndex + 1
				} else {
					next.hintDoneIndex = -1
					next.hints = []
				}
				return next
			})
		}
		return [res, commit, err]
	}
}
export type HttpStateStore = {
	loading: Record<ApiMethodKeys, number>
	errors: Record<ApiMethodKeys, null | Error | ConnectError>
}

export const storeHandler: ApiType = new ApiStore()
export const httpStateStore = writable<HttpStateStore>({
	loading: objectKeys(storeHandler).reduce(
		(r, k) => ({ ...r, [k]: 0 }),
		{} as HttpStateStore['loading']
	),
	errors: objectKeys(storeHandler).reduce(
		(r, k) => ({ ...r, [k]: null }),
		{} as HttpStateStore['errors']
	),
})

type GoFunc<Params, Result> = (params: Params) => GoResult<Result>
type GoResult<Result> = Promise<[Result, null] | [null, Result]>
type CommitableGoResult<Result> = Promise<
	| [result: Result, commit: () => Promise<void>, error: null]
	| [result: null, commit: null, error: ConnectError | Error]
>
