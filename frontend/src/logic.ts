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
