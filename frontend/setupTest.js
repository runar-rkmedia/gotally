import.meta.env.SSR = false
import matchers from '@testing-library/jest-dom/matchers';
import { expect, vi } from 'vitest';
import createFetchMock from 'vitest-fetch-mock';
import { rest } from 'msw'
import { setupServer } from 'msw/node'

const fetchMock = createFetchMock(vi)
// fetchMock.enableMocks()
vi.mock('$app/env.js', () => ({
  amp: false,
  browser: true, // or false for testing ssr
  dev: true,
  mode: 'test'
}))
expect.extend(matchers);
const encode = (obj) => {
  return Buffer.from(JSON.stringify(obj)).toString('base64')
}
const restHandlers = [

  rest.all(/.*/, async (req, res, ctx) => {
    const extraHeaders = {
      DEV_GAME_OPTIONS: encode({ seed: 123, state: 123 }),
      DEV_USERNAME: 'Vitest',
      Authorization: ('vitest-' + Math.random() * 1e6).slice(0, 21)
    }
    console.log('resty')
    for (const [k, v] of Object.entries(extraHeaders)) {
      req.headers.set(k, v)
    }
    console.log('headers', req.headers)
    const originalResponse = await (await ctx.fetch(req)).json()
    // return originalResponse
    return res(ctx.json(originalResponse))
    // return req.passthrough()
  })
]

const server = setupServer(...restHandlers)


// Start server before all tests
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))

//  Close server after all tests
afterAll(() => server.close())

// Reset handlers after each test `important for test isolation`
afterEach(() => server.resetHandlers())
