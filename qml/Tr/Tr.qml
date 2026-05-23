pragma Singleton
import QtQuick

QtObject {
    function i18n(message, arg1, arg2, arg3) {
        var result = message
        if (arguments.length > 1 && arg1 !== undefined) {
            result = result.replace("%1", arg1)
        }
        if (arguments.length > 2 && arg2 !== undefined) {
            result = result.replace("%2", arg2)
        }
        if (arguments.length > 3 && arg3 !== undefined) {
            result = result.replace("%3", arg3)
        }
        return result
    }
}
