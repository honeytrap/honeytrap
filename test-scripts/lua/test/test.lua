--
-- Created by IntelliJ IDEA.
-- User: nordi
-- Date: 12-6-2018
-- Time: 14:21
-- To change this template use File | Settings | File Templates.
--
function canHandle(message)
    if (message == "pass") then
        return true
    elseif (message == "fail") then
        return false
    end

    return true
end

function handle(message)
    if message == "logCritical" then
        doLog("critical", "critical")
    elseif message == "logDebug" then
        doLog("debug", "debug")
    elseif message == "logError" then
        doLog("error", "error")
    elseif message == "logFatal" then
        doLog("fatal", "fatal")
    elseif message == "logInfo" then
        doLog("info", "info")
    elseif message == "logNotice" then
        doLog("notice", "notice")
    elseif message == "logPanic" then
        doLog("panic", "panic")
    elseif message == "logWarning" then
        doLog("warning", "warning")
    end

    if parameterTest ~= nil then
        local response = parameterTest("key", "value")
        if response == nil then
            return message
        else
            return message .. parameterTest("key", "value")
        end
    end

    return message
end