// BEGIN EXTRA CODE
// END EXTRA CODE
/**
 * Opens a navigation application on your device or a web browser showing Google Maps directions.
 * @param {string} location - This field is required.
 * @returns {Promise.<boolean>}
 */
async function NavigateTo(location) {
    // BEGIN USER CODE
    if (!location) {
        return Promise.reject(new Error("Input parameter 'Location' is required"));
    }
    location = encodeURIComponent(location);
    const iosUrl = `maps://maps.apple.com/?daddr=${location}`;
    const androidUrl = `google.navigation:q=${location}`;
    const webUrl = `https://maps.google.com/maps?daddr=${location}`;
    // Native platform
    if (navigator && navigator.product === "ReactNative") {
        const Linking = require("react-native").Linking;
        const Platform = require("react-native").Platform;
        const url = Platform.select({
            ios: iosUrl,
            default: androidUrl
        });
        return Linking.canOpenURL(url).then(supported => {
            if (!supported) {
                return false;
            }
            return Linking.openURL(url).then(() => true);
        });
    }
    // Hybrid or mobile web platform
    if (window && window.navigator.userAgent) {
        // iOS platform
        if (/iPad|iPhone|iPod/i.test(window.navigator.userAgent)) {
            openUrl(iosUrl);
            return Promise.resolve(true);
        }
        // Android platform
        if (/android|sink/i.test(window.navigator.userAgent)) {
            openUrl(androidUrl);
            return Promise.resolve(true);
        }
    }
    // Desktop web or other platform
    if (window) {
        window.location.href = webUrl;
        return Promise.resolve(true);
    }
    return Promise.resolve(false);
    function openUrl(url) {
        // Hybrid platform
        if (window && window.cordova) {
            window.open(url, "_system");
        }
        // Mobile web platform
        if (window) {
            window.location.href = url;
        }
    }
    // END USER CODE
}

export { NavigateTo };
