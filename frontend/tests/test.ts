import { test, expect, type Page, type Locator } from '@playwright/test'

// Coming from Cypress, Playwright seems a bit lacking. I am probably using it wrong.
//

test('play a short game', async ({ page }) => {
	// Go to http://localhost:5173/

	await page.goto('http://localhost:5173/')
	await page.evaluate(() =>
		localStorage.setItem('userSettings', JSON.stringify({ swipeAnimationTime: 0 }))
	)
	await page.locator('.board').click()
	const swipe = async (page: Page, direction: 'Up' | 'Right' | 'Down' | 'Left') => {
		await Promise.all([boardKey(page, 'Arrow' + direction), waitForResponse(page, /Swipe/)])
	}
	expect(await getScore(page)).toBe('0')
	await swipe(page, 'Right')
	await swipe(page, 'Up')
	// Combine 8 and 8
	await clickCellCoord(page, 5, 1)
	await clickCellCoord(page, 5, 2)
	// await clickCellCoord(page, 5, 2)
	await evaluate(page)
	expect(await getScore(page)).toBe('16')
	await swipe(page, 'Up')
	await swipe(page, 'Right')
	await swipe(page, 'Down')
	// await page.locator('button', { hasText: 'Hint' }).click()
	const getHint = async (page: Page) => {
		await Promise.all([boardKey(page, 'h'), waitForResponse(page, /Hint/)])
	}
	await getHint(page)
	await page.pause()
	// There should now be a single hint
	await page.waitForSelector('.hinted')
	const hintedCount = await page.locator('.hinted').count()
	// const hinted = page.locator('.hinted')
	expect(hintedCount).toBe(4)
	const vals = await values(page.locator('.hinted'))
	expect(vals.join(';')).toBe('16;1;8;2')
	// Combine all the items from the hint
	await clickCellIndex(page, 24)
	await boardKey(page, 'ArrowRight')
	await boardKey(page, 'ArrowUp')
	await boardKey(page, 'ArrowUp')
	await evaluate(page)
	expect(await getScore(page)).toBe('48')
	await page.reload()
	await boardKey(page, 'ArrowUp')
	await clickCellIndex(page, 2)
	await clickCellIndex(page, 3)
	await evaluate(page)
	expect(await getScore(page)).toBe('50')
	expect(await getMoves(page)).toBe('9')
	await page.reload()

	expect(await getMoves(page), 'Moves should not change on page-reload').toBe('9')
	expect(await getScore(page), 'Score should not change on page-reload').toBe('50')
})
const evaluate = async (page: Page) => {
	await boardKey(page, ' ')
	return waitForResponse(page, /CombineCells/)
}

const getScore = async (page: Page) => {
	return page.locator('.score').getAttribute('data-score')
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
	await wait(page)
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
	await wait(page)
	return c
}
const wait = (page: Page) => Promise.resolve() // page.waitForTimeout(waitTime)

export const values = async (els: Locator) => {
	const v = await els.locator('.cellValue').allInnerTexts()
	return v
}
export const coordToIndex = (y: number, x: number, maxColumns: number, maxRows: number) => {
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
