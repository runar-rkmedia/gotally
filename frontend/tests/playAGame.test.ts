import { test, expect } from '@playwright/test'
import { testUtil } from 'utils'
const {
	setup,
	swipe,
	clickCellCoord,
	getHint,
	evaluate,
	getScore,
	values,
	clickCellIndex,
	boardKey,
	getMoves
} = testUtil

// Coming from Cypress, Playwright seems a bit lacking. I am probably using it wrong.
//

test('play a short game', async ({ page }, testInfo) => {
	// Go to http://localhost:5173/
	await setup(page)
	await page.locator('.board').click()
	expect(await getScore(page)).toBe('0')
	const initialBoard = await page.locator('.board').first().innerText()
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
	await getHint(page)
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

	// Restart the game
	await page.locator('[data-testid="menu"]').click()
	await page.locator('[data-test-id="restart-game"]').click()
	await page.waitForLoadState('networkidle')
	expect(await getScore(page)).toBe('0')
	expect(await getMoves(page)).toBe('0')
	const board = await page.locator('.board').first().innerText()
	expect(board).toBe(initialBoard)
	const scr = await page.screenshot({ fullPage: true })
	await testInfo.attach('screenshot', { body: scr, contentType: 'image/png' })
})
