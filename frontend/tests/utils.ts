import type { Page, Locator } from '@playwright/test'

const evaluate = async (page: Page) => {
	await boardKey(page, ' ')
	return waitForResponse(page, /CombineCells/)
}
const setup = async (page: Page) => {
	page.on('response', async (res) => {
		const statusCode = res.status()
		const req = res.request()
		if (statusCode >= 400) {
			console.log(
				'Response',
				statusCode,
				String(await res.body()),
				await res.json(),
				req.method(),
				req.url(),
				req.postDataJSON()
			)
			throw new Error('Api returned ${statusCode}')
		}
	})
	await page.goto('http://localhost:5173')
	await page.evaluate(() =>
		localStorage.setItem('userSettings', JSON.stringify({ swipeAnimationTime: 0 }))
	)
}
const swipe = async (page: Page, direction: 'Up' | 'Right' | 'Down' | 'Left') => {
	await Promise.all([boardKey(page, 'Arrow' + direction), waitForResponse(page, /Swipe/)])
}
const getHint = async (page: Page) => {
	await Promise.all([boardKey(page, 'h'), waitForResponse(page, /Hint/)])
}

const locateTestElement = (page: Page, testId = '') => {
	return page.locator(`[data-testid="${testId}"]`)
}
const getScore = async (page: Page) => {
	return locateTestElement(page, 'score').getAttribute('data-score')
}
const getMoves = async (page: Page) => {
	return page.locator('.moves').getAttribute('data-moves')
}
const waitForResponse = async (page: Page, path: RegExp) => {
	console.debug('waiting for response on: ', path)
	const response = await page.waitForResponse(path, { timeout: 1500 })
	console.debug('got response', path)
	return response
}

const boardKey = async (page: Page, key: string) => {
	console.log('boardKey', key)
	await page.locator('.board').press(key)
	// await wait(page)
}
const clickCellCoord = async (page: Page, row: number, column: number, columns = 5, rows = 5) => {
	console.log('selecting', row, column)
	const index = coordToIndex(column, row, columns, rows)
	if (!index) {
		console.error('failed to create index for coor', { column, row, columns, rows })
		return
	}
	return clickCellIndex(page, index)
}
const clickCellIndex = async (page: Page, index: number) => {
	const selector = `.cell:nth-of-type(${index})`
	console.log('selecting index', index, selector)
	const c = page.locator(selector).click()
	// await expect(page.locator(selector)).toHaveClass('selected')
	// await wait(page)
	return c
}
// const wait = (_page: Page) => Promise.resolve() // page.waitForTimeout(waitTime)

const values = async (els: Locator) => {
	const v = await els.locator('.cellValue').allInnerTexts()
	return v
}
const coordToIndex = (y: number, x: number, maxColumns: number, maxRows: number) => {
	if (x < 0) {
		return null
	}
	if (y < 0) {
		return null
	}
	if (y > maxRows) {
		return null
	}
	if (x > maxColumns) {
		return null
	}

	return y * maxColumns + x
}
export const testUtil = {
	setup,
	locateTestElement,
	swipe,
	clickCellCoord,
	getHint,
	evaluate,
	getScore,
	values,
	clickCellIndex,
	boardKey,
	getMoves
}
