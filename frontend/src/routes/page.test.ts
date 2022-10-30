import { render, screen, waitFor } from '@testing-library/svelte'
import Index from './+page.svelte'

// About testing:
// Testing are encouraged to target a local development server instead of mocking it.
// Therefore, this test does not mock anything.
//
// TODO: add tests

describe('Test index.svelte', () => {
	it('link to svelte website', async () => {
		render(Index)
		await waitFor(() => screen.getByText(/Score/))
		// Game should now be ready
		const username = await screen.findByText(/Username:/)
		expect(username).toHaveTextContent('Vitest')
	})
})
