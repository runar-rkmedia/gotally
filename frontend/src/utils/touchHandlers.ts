import { findDOMParent } from './findDomParent'

export const findCellFromTouch = (e: TouchEvent): [Finding, null] | [null, FindingError] => {
	const touch = e.changedTouches?.item?.(0)
	if (!touch) {
		return [
			null,
			{
				code: 'no-touch',
				message: 'failed to find the current item from the touch',
				details: { e },
			},
		]
	}
	const x = touch.pageX
	const y = touch.pageY
	const target = document.elementFromPoint(x, y)
	if (!target) {
		return [
			null,
			{
				code: 'no-target',
				message: 'failed to find the target from the touch',
				details: { e, touch },
			},
		]
	}
	const cell = findDOMParent(target, (e) => e.classList.contains('cell')) as HTMLDivElement
	if (!cell) {
		return [
			null,
			{
				code: 'no-cell',
				message: 'failed to find the cell from the touch',
				details: { e, touch, target },
			},
		]
	}
	const isEmpty = cell.classList.contains('blank')
	if (!cell.parentElement?.classList.contains('board')) {
		return [
			null,
			{
				code: 'no-parent-board',
				message: 'failed to find the board from the touch',
				details: { e, touch, cell, target, isEmpty },
			},
		]
	}
	const index = [...(cell.parentElement?.children || [])].findIndex((el) => el === cell)
	if (index === -1) {
		return [
			null,
			{
				code: 'no-index',
				message: 'failed to find the index from the touch',
				details: { e, touch, cell, target, isEmpty },
			},
		]
	}
	return [{ cell, index, target, touch, x, y, e, isEmpty }, null]
}

type Finding = {
	e: TouchEvent
	touch: Touch
	target: Element
	index: number
	cell: HTMLDivElement
	x: number
	y: number
	isEmpty: boolean
}

type FindingError = {
	message: string
	code: 'no-touch' | 'no-target' | 'no-cell' | 'no-index' | 'no-parent-board'
	details: Partial<Finding>
}
