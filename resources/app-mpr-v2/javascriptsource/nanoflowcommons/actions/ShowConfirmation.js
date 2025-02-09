// BEGIN EXTRA CODE
// END EXTRA CODE
/**
 * Shows a confirmation dialog during the execution of a nanoflow, to make perform actions based on the user input.
 * @param {string} titleCaption - Set to empty to use default text 'Confirmation'. (Only for native)
 * @param {string} question - This field is required.
 * @param {string} cancelButtonCaption - Set to empty to use default text 'Cancel'.
 * @param {string} proceedButtonCaption - Set to empty to use default text 'OK'.
 * @returns {Promise.<boolean>}
 */
async function ShowConfirmation(titleCaption, question, cancelButtonCaption, proceedButtonCaption) {
    // BEGIN USER CODE
    if (!question) {
        return Promise.reject(new Error("Input parameter 'Question' is required"));
    }
    const cancel = cancelButtonCaption || "Cancel";
    const proceed = proceedButtonCaption || "OK";
    const title = titleCaption || "Confirmation";
    // Native platform
    if (navigator && navigator.product === "ReactNative") {
        const Alert = require("react-native").Alert;
        return new Promise(resolve => {
            Alert.alert(title, question, [
                { text: cancel, onPress: () => resolve(false), style: "cancel" },
                { text: proceed, onPress: () => resolve(true) }
            ]);
        });
    }
    // Other platforms
    return new Promise(resolve => {
        mx.ui.confirmation({
            content: question,
            proceed,
            cancel,
            handler: () => resolve(true),
            onCancel: () => resolve(false)
        });
    });
    // END USER CODE
}

export { ShowConfirmation };
