import {
	createPromiseClient,
	createConnectTransport,
	type Interceptor,
	type ConnectTransportOptions,
	Code,
} from '@bufbuild/connect-web'
import { BoardService } from './'
import { ConnectError } from '@bufbuild/connect-web'
import * as appEnv from '$app/environment'
import { writable } from 'svelte/store'
import type { ApiType } from './store'

const state = {
	authHeader: appEnv.browser && localStorage.getItem('sessionID'),
}

type ErrorStore = {
	errors: Array<{
		error: Error | ConnectError
		time: Date
		url: string
	}>
}

export const httpErrorStore = writable<ErrorStore>({ errors: [] })

const retrier: Interceptor =
	(next, retries = 0) =>
	async (req) => {
		try {
			if (state.authHeader) {
				req.header.set('Authorization', state.authHeader)
			}
			const res = await next(req)
			if (appEnv.browser) {
				const authHeader = res.header.get('Authorization')
				if (authHeader) {
					localStorage.setItem('sessionID', authHeader)
					state.authHeader = authHeader
				}
			}

			if (res.stream) {
				console.warn('got a streaming response, not implemented')
				return {
					...res,

					async read() {
						const msg = await res.read()
						console.debug('streaming response', msg)
						return msg
					},
				}
			}
			httpErrorStore.update((e) => ({
				...e,
				errors: e.errors.filter((err) => err.url === req.url),
			}))
			return res
		} catch (err: unknown) {
			if (err instanceof Error) {
				httpErrorStore.update((e) => ({
					...e,
					errors: [{ time: new Date(), error: err as any, url: req.url }],
				}))
				setTimeout(() => {
					const cutoff = new Date().getTime() - 10000
					httpErrorStore.update((e) => ({
						...e,
						errors: e.errors.filter((err) => err.time.getTime() > cutoff),
					}))
				}, 10500)
				if (err.message.includes('NetworkError')) {
					const waitPeriod = Math.min(100 * retries, 3000)
					await new Promise((r) => setTimeout(r, waitPeriod))
					return (retrier as any)(next, ++retries)(req)
				} else if (err instanceof ConnectError) {
					switch (err.code) {
						case Code.Unauthenticated: {
							if (appEnv.browser) {
								localStorage.removeItem('sessionID')
								state.authHeader = ''
								req.header.delete('Authorization')
								if (retries < 5) {
									return (retrier as any)(next, ++retries)(req)
								}
							}
						}
					}
				}
			}

			throw err
		}
	}

const isHttps = appEnv.browser && document.location.protocol.includes('https')
const isTest = (appEnv as any).mode === 'test'
const transportOptions: ConnectTransportOptions = {
	baseUrl:
		import.meta.env?.VITE_API ||
		(isHttps
			? '/'
			: import.meta.env?.VITE_DEV_API ||
			  `http://${appEnv.browser ? window.location.hostname : 'localhost'}:8080/`),
	interceptors: [retrier],
	useBinaryFormat: isTest
		? false
		: appEnv.browser
		? !window.location.search.includes('json=1')
		: true,
}
console.debug({ transportOptions, meta: import.meta, appEnv })

const transport = createConnectTransport(transportOptions)

export const client = createPromiseClient(BoardService, transport)

export const handleError = (key: keyof ApiType, err: ConnectError | Error) => {
	// TODO: handle error
	if (err instanceof ConnectError) {
		const { metadata, ...all } = err
		const meta: any = {}
		for (const [k, v] of metadata) {
			meta[k] = v
		}
		console.warn(`[${err.code}]: ${err.name} - ${err.message}`, all, meta)
	}
	console.warn('An error occured', err)
}

/** go-like error-handling, instaed of throwing. 

takes in any promise never throws, but instead returns a tuple with 
either the result in the first item, or an error in the second.
*/
export const go = async <P = any>(promise: Promise<P>): GoPromise<P> => {
	try {
		const result = await promise
		return [result, null]
	} catch (err) {
		console.dir(err, { depth: null })
		return [null, err as any]
	}
}

export type GoPromise<P> = Promise<[P, null] | [null, ConnectError | Error]>
