// jest-dom adds custom matchers for asserting on DOM nodes, e.g.
// expect(element).toHaveTextContent(/react/i)
// https://github.com/testing-library/jest-dom
import '@testing-library/jest-dom/vitest';

// jsdom has no matchMedia. The library's own theme code guards for that, but
// Mantine calls it unguarded, so stand one in.
window.matchMedia = window.matchMedia || ((query) => ({
  media: query,
  matches: false,
  onchange: null,
  addListener: () => { },
  removeListener: () => { },
  addEventListener: () => { },
  removeEventListener: () => { },
  dispatchEvent: () => false,
}))
