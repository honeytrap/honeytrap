local json = require "generic/util/json"
-- Test script for Lua functions

function canHandle()
    return true
end

function handle(message)
    if (message == "EOF") then
        return "_return"
    end

    if (message == "test") then
        return "test"
    end

    local request = getRequest("1")
    request = json.parse(request)

    local body = request.body

    if (body.username == "test" and body.password == "test") then
        restWrite("200", [[{"login": "success"}]], [[{}]])
    else
        restWrite("200", [[{"login": "failed"}]], [[{}]])
    end


    return "_return"
end
