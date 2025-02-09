// BEGIN EXTRA CODE
// END EXTRA CODE
/**
 * Hides a loading dialog.
 * @param {Big} identifier - This field is required.
 * @returns {Promise.<void>}
 */
async function HideProgress(identifier) {
    // BEGIN USER CODE
    if (identifier == null) {
        return Promise.reject(new Error("Input parameter 'Identifier' is required"));
    }
    mx.ui.hideProgress(Number(identifier));
    return Promise.resolve();
    // END USER CODE
}

export { HideProgress };
