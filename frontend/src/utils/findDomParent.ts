export function findDOMParent<E extends Element>(
	element: E,
	predecate: (element: E) => boolean
): E | null {
	if (!element) {
		return null
	}
	const result = predecate(element)
	if (result) {
		return element
	}
	if (!element.parentElement) {
		return null
	}
	return findDOMParent(element.parentElement as any, predecate)
}
