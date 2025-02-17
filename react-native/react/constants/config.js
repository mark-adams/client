/* @flow */

export type NavState = 'navStartingUp' | 'navNeedsRegistration' | 'navNeedsLogin' | 'navLoggedIn' | 'navErrorStartingUp'

// Constants
export const navStartingUp = 'navStartingUp'
export const navNeedsRegistration = 'navNeedsRegistration'
export const navNeedsLogin = 'navNeedsLogin'
export const navLoggedIn = 'navLoggedIn'
export const navErrorStartingUp = 'navErrorStartingUp'

// Actions
export const startupLoading = 'config:startupLoading'
export const startupLoaded = 'config:startupLoaded'

export const statusLoaded = 'config:statusLoaded'
export const configLoaded = 'config:configLoaded'

export const devConfigLoading = 'config:devConfigLoading'
export const devConfigLoaded = 'config:devConfigLoaded'
export const devConfigUpdate = 'config:devConfigUpdate'
export const devConfigSaved = 'config:devConfigSaved'
