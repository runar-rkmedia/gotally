import { cellValue } from './components/board/cell'
import type { Cell } from './connect-web'

const getRows = (boardEl: HTMLElement, nRows: number) => {
	const cells = [...boardEl.children] as HTMLElement[]
	const rows = new Array(nRows).fill(null).map((_, i) => cells.slice(i * nRows, i * nRows + nRows))
	return rows
}
const getColumns = (boardEl: HTMLElement, nColumns: number) => {
	const cells = [...boardEl.children] as HTMLElement[]

	const columns: HTMLElement[][] = new Array(nColumns).fill(null).map(() => [])
	for (let i = 0; i < columns.length; i++) {
		for (let j = 0; j < columns.length; j++) {
			;(columns as any)[i][j] = cells[j * nColumns + i]
		}
	}
	return columns
}

const cellIsEmpty = (cell: Element) =>
	!cell || !cell.hasChildNodes() || cell.classList.contains('blank')

/** Gets the offset of an element with subpixel-preciction */
const cumulativeOffset = function (el: HTMLElement) {
	let element = el
	let top = 0,
		left = 0
	do {
		top += element.getBoundingClientRect().y || 0
		left += element.getBoundingClientRect().x || 0
		element = element.offsetParent as any
	} while (element)

	return {
		top: top,
		left: left
	}
}

export const animateSwipe = async ({
	boardEl,
	nColumns,
	nRows,
	positive,
	vertical,
	swipeAnimationTime,
	dry
}: {
	boardEl: HTMLElement
	nColumns: number
	nRows: number
	positive: boolean
	vertical: boolean
	swipeAnimationTime: number
	dry?: boolean
}) => {
	return new Promise((res) => {
		const rows = vertical ? getColumns(boardEl, nColumns) : getRows(boardEl, nRows)
		let didAnimate = false
		for (const [r, row] of rows.entries()) {
			let empties = 0
			for (
				let i = positive ? row.length - 1 : 0;
				positive ? i > -1 : i < row.length;
				i += positive ? -1 : 1
			) {
				const nextIndex = positive ? i + 1 : i - 1
				const current = row[i]
				if (!current) {
					console.warn('cell was undefined', {
						r,
						row,
						i,
						rows,
						boardLength: boardEl.children.length,
						boardChildren: boardEl.children
					})
					return false
				}

				if (cellIsEmpty(current)) {
					empties++
					continue
				}
				if (!empties) {
					continue
				}
				const next = row[nextIndex]
				if (!next) {
					continue
				}
				if (dry) {
					res(true)
					return
				}
				didAnimate = true
				// state.didAnimate = true
				// state.isAnimating = true
				const targetIndex = positive ? i + empties : i - empties
				const target = row[targetIndex]

				const pos = cumulativeOffset(current)
				const posNext = cumulativeOffset(target)
				if (vertical) {
					const diffY = posNext.top - pos.top
					current.style.transform = `translateY(${diffY}px)`
				} else {
					const diffX = posNext.left - pos.left
					current.style.transform = `translateX(${diffX}px)`
				}
				current.style.transition = `transform ${swipeAnimationTime}ms var(--easing-standard)`
			}
		}
		if (didAnimate) {
			setTimeout(() => res(didAnimate), swipeAnimationTime)
			return
		}
		return res(didAnimate)
	})
}

export const coordToIndex = (x: number, y: number, maxColumns: number, maxRows: number) => {
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

type invalidPathError = {
	message: string
	invalidIndex: number
}

const newInvalidPathError = (message: string, invalidIndex: number): invalidPathError => ({
	message,
	invalidIndex
})

type cellLike = Cell | { base: number; twopow: number }

export const ValidatePath = (
	indexes: number[],
	rows: number,
	columns: number,
	cells: cellLike[]
) => {
	const nIndexes = indexes.length
	if (nIndexes <= 1) {
		return newInvalidPathError('path is too short', 0)
	}
	const nCells = cells.length
	if (nIndexes > nCells) {
		return newInvalidPathError('path is too long', 0)
	}
	const seen: Record<number, number> = {}
	let prevIndex = -1
	for (const [i, index] of indexes.entries()) {
		if (seen[index] !== undefined) {
			return newInvalidPathError(
				`duplicate entry for index at position ${index} /${seen[index]}`,
				i
			)
		}
		if (index > nCells) {
			return newInvalidPathError(`ErrPathIndexOutsideBounds for index ${index} at position ${i}`, i)
		}
		if (index < 0) {
			return newInvalidPathError(`ErrPathIndexOutsideBounds for index ${index} at position ${i}`, i)
		}
		const cell = cells[index]
		const cellV = cellValue(cell)
		if (!cellV) {
			return newInvalidPathError(`ErrPathIndexEmptyCell for index ${index} at position ${i}`, i)
		}
		if (prevIndex >= 0 && !AreNeighboursByIndex(index, prevIndex, columns, rows)) {
			return newInvalidPathError(`Not a neighbour ${index} ${prevIndex}`, i)
		}
		seen[index] = i
		prevIndex = index
	}
	return null
}

export const AreNeighboursByIndex = (
	a: number,
	b: number,
	columns: number,
	rows: number
): boolean => {
	if (a === b) {
		return false
	}
	if (a < 0 || b < 0) {
		return false
	}
	const max = columns * rows
	if (a >= max || b >= max) {
		return false
	}
	const [ac, ar] = IndexToCord(a, columns)
	const [bc, br] = IndexToCord(b, columns)

	const diffc = ac - bc
	const diffr = ar - br

	switch (true) {
		// The cells cannot both be on different columns and rows and still be neighbours
		case diffc !== 0 && diffr !== 0:
			return false
		// The cells cannot be the same
		case diffc === 0 && diffr === 0:
			return false
		case diffc == 1:
			return true
		case diffc == -1:
			return true
		case diffr == 1:
			return true
		case diffr == -1:
			return true
	}
	return false
}

function cellRow(i: number, columns: number) {
	return Math.floor(i / columns)
}
function cellColumn(i: number, columns: number) {
	return i % columns
}
function IndexToCord(i: number, columns: number) {
	return [cellColumn(i, columns), cellRow(i, columns)] as const
}
