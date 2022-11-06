import { assert, describe, it } from 'vitest'
import { AreNeighboursByIndex } from './logic'

describe('Check AreNeightbours-logic', () => {
	const notNeighbours = [[8, 17, 5, 5]]
	const neighbours = [[5, 10, 5, 5]]
	it('AreNeighboursByIndex', () => {
		for (const [a, b, columns, rows] of neighbours) {
			assert.isTrue(
				AreNeighboursByIndex(a, b, columns, rows),
				`${a} and ${b} should not be neighbours on a ${columns}x${rows}-board`
			)
		}
		for (const [a, b, columns, rows] of notNeighbours) {
			assert.isFalse(
				AreNeighboursByIndex(a, b, columns, rows),
				`${a} and ${b} should not be neighbours on a ${columns}x${rows}-board`
			)
		}
	})
})
