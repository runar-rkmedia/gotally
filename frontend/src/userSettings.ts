import { writable } from 'svelte/store'
import { browser } from '$app/env'

const getLocalStorageJson = <T>(key: string): T | null => {
	if (!browser) {
		return null
	}
	const value = localStorage.getItem(key)
	if (!value) {
		return null
	}
	try {
		return JSON.parse(value) as T
	} catch (error) {
		console.error('failed to read from localStorage', key, error)
		return null
	}
}

type UserSettings = {
	/** Animation-time for swipes */
	swipeAnimationTime: number
}

const userSettings = writable(
	getLocalStorageJson<UserSettings>('userSettings') || { swipeAnimationTime: 480 }
)

export default userSettings
