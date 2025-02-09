// BEGIN EXTRA CODE
// END EXTRA CODE
/**
 * Start drafting an email in the platform specified email client. This might work differently for each user depending on their platform and local configuration.
 * @param {string} recipient - The recipient, or recipients, separated by comma's.
 * @param {string} cc - The Carbon Copy recipient, or recipients, separated by comma's.
 * @param {string} bcc - The Blind Carbon Copy recipient, or recipients, separated by comma's.
 * @param {string} subject
 * @param {string} body
 * @returns {Promise.<boolean>}
 */
async function DraftEmail(recipient, cc, bcc, subject, body) {
    // BEGIN USER CODE
    let url = "mailto:";
    if (recipient) {
        url += `${encodeURI(recipient)}?`;
    }
    if (cc) {
        url += `cc=${encodeURIComponent(cc)}&`;
    }
    if (bcc) {
        url += `bcc=${encodeURIComponent(bcc)}&`;
    }
    if (subject) {
        url += `subject=${encodeURIComponent(subject)}&`;
    }
    if (body) {
        url += `body=${encodeURIComponent(body)}&`;
    }
    // Remove the last '?' or '&'
    url = url.slice(0, -1);
    // Native platform
    if (navigator && navigator.product === "ReactNative") {
        const Linking = require("react-native").Linking;
        return Linking.canOpenURL(url).then(supported => {
            if (!supported) {
                return false;
            }
            return Linking.openURL(url).then(() => true);
        });
    }
    // Hybrid platform
    if (window && window.cordova) {
        window.open(url, "_system");
        return Promise.resolve(true);
    }
    // Web platform
    if (window) {
        window.location.href = url;
        return Promise.resolve(true);
    }
    return Promise.resolve(false);
    // END USER CODE
}

export { DraftEmail };
