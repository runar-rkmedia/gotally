import type { Cell } from '../../connect-web'

export const cellValueAsString = (c: Cell | { base: number; twopow: number }) => {
	const n = cellValue(c)
	if (!n) {
		return ''
	}
	return 'n'
}
export const cellValue = (c: Cell | { base: number; twopow: number }) => {
	if (!c) {
		return 0
	}
	if (Number(c.base) === 0) {
		return 0
	}
	return Number(c.base) * Math.pow(2, Number(c.twopow))
}
export const primeFactors = (n: number) => {
	switch (n) {
		case 0:
			return []
		case 1:
		case 2:
		case 5:
		case 7:
		case 11:
			return [n]
	}

	const factors: Array<number> = []
	let primeIndex = 0
	while (primeIndex < primeCount) {
		const prime = firstThousandPrimes[primeIndex]
		if (prime === n) {
			factors.push(prime)
			return factors
		}
		const mod = n % prime
		if (mod > 0) {
			primeIndex++
			continue
		}
		n /= prime
		factors.push(prime)
	}
	if (!factors.length) {
		throw new Error(`Failed to find primefactors for ${n}. Need more primes.`)
	}
	return factors
}

/** List of primes, up to 997. Highly unlikely we will ever need more. (We probably only need the first five)
 */
const firstThousandPrimes = [
	2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97,
	101, 103, 107, 109, 113, 127, 131, 137, 139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193,
	197, 199, 211, 223, 227, 229, 233, 239, 241, 251, 257, 263, 269, 271, 277, 281, 283, 293, 307,
	311, 313, 317, 331, 337, 347, 349, 353, 359, 367, 373, 379, 383, 389, 397, 401, 409, 419, 421,
	431, 433, 439, 443, 449, 457, 461, 463, 467, 479, 487, 491, 499, 503, 509, 521, 523, 541, 547,
	557, 563, 569, 571, 577, 587, 593, 599, 601, 607, 613, 617, 619, 631, 641, 643, 647, 653, 659,
	661, 673, 677, 683, 691, 701, 709, 719, 727, 733, 739, 743, 751, 757, 761, 769, 773, 787, 797,
	809, 811, 821, 823, 827, 829, 839, 853, 857, 859, 863, 877, 881, 883, 887, 907, 911, 919, 929,
	937, 941, 947, 953, 967, 971, 977, 983, 991, 997
] as const
const primeCount = firstThousandPrimes.length
// const numbers = [
// 	1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 53, 64, 80, 125, 500, 210, 210, 210, 210, 210, 210, 210, 210, 420
// ]
// // const numbers = [12]

// for (const n of numbers) {
// 	console.log(`number ${n}: ${primeFactors(n)}`)
// }
