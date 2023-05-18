import { assert, describe, it } from 'vitest'
import { AreNeighboursByIndex, coordToIndex, indexToCoord } from './logic'

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
describe('index to coordinate', () => {
	const tests = [
		[0, [0, 0]],
		[2, [2, 0]],
		[3, [0, 1]],
		[4, [1, 1]],
		[5, [2, 1]],
		[6, [0, 2]],
	] as const
	it('coordinate for index should match for ', () => {
		for (const [i, want] of tests) {
			const coord = indexToCoord(i, 3)
			assert.equal(
				coord.join(','),
				want.join(','),
				`For index=${i}, expected ${want.join(',')} but got ${coord.join(',')} `
			)
		}
	})
})
describe('Coordinate to index', () => {
	it(`coordinate to index and back should return same value`, () => {
		const testRows = new Array(3).fill(null).map((_, i) => i + 3)
		const testColumns = new Array(3).fill(null).map((_, i) => i + 3)
		for (const rows of testRows) {
			for (const columns of testColumns) {
				const testIndexes = new Array(rows * columns).fill(null).map((_, i) => i)
				for (const i of testIndexes) {
					const [x, y] = indexToCoord(i, columns)
					const index = coordToIndex(x, y, columns, rows)
					// assert.isNotNull(index)
					assert.equal(
						index,
						i,
						`Expected indexToCoord(${i}, ${columns}, ${rows}) and back with coordToIndex(${x}, ${y}, ${columns}, ${rows}) to equal input of ${i}, but was ${index}`
					)
				}
			}
		}
	})
})
