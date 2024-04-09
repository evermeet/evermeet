
import Package from '../../../../package.json';
import { apiCall } from './api.js';

let loadedConfig;

// load configuration
export async function loadConfig () {
    if (!loadedConfig) {
        loadedConfig = await apiCall(fetch, 'config');
    }
    return loadedConfig
}

export const pkg = Package;