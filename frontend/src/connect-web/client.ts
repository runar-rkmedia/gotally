import {
	createPromiseClient,
	createConnectTransport,
	type Interceptor,
	type ConnectTransportOptions
} from '@bufbuild/connect-web'
import { BoardService } from './'
import { ConnectError } from '@bufbuild/connect-web'
import { browser } from '$app/env'

const state = {
	authHeader: browser && localStorage.getItem('sessionID')
}

const retrier: Interceptor =
	(next, retries = 0) =>
	async (req) => {
		try {
			if (state.authHeader) {
				req.header.set('Authorization', state.authHeader)
			}
			const res = await next(req)
			if (browser) {
				const authHeader = res.header.get('Authorization')
				if (authHeader) {
					localStorage.setItem('sessionID', authHeader)
					state.authHeader = authHeader
				}
			}

			if (res.stream) {
				console.log('got a streaming response')
				return {
					...res,

					async read() {
						const msg = await res.read()
						console.log('streaming response', msg)
						return msg
					}
				}
			}
			return res
		} catch (err) {
			if (err instanceof Error) {
				if (err.message.includes('NetworkError')) {
					const waitPeriod = Math.min(100 * retries, 3000)
					await new Promise((r) => setTimeout(r, waitPeriod))
					return (retrier as any)(next, ++retries)(req)
				}
			}

			throw err
		}
	}

const isHttps = browser && document.location.protocol.includes('https')
const transportOptions: ConnectTransportOptions = {
	baseUrl: import.meta.env.VITE_API || (isHttps ? '/' : 'http://localhost:8080/'),
	interceptors: [retrier],
	useBinaryFormat: false
}
console.debug(import.meta.env, { transportOptions })

const transport = createConnectTransport(transportOptions)

export const client = createPromiseClient(BoardService, transport)

export const handleError = (err: ConnectError | Error) => {
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
export const go = async <P = any>(
	promise: Promise<P>
): Promise<[P, null] | [null, ConnectError | Error]> => {
	try {
		const result = await promise
		return [result, null]
	} catch (err) {
		console.dir(err, { depth: null })
		return [null, err as any]
	}
}
