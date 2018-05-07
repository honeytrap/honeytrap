--- Define whether the Lua script can handle the request
-- canHandle function returns the approval or disapproval to handle the connection. This is based on the information
-- of the current connection and new request. The service is already determined by the Honeytrap implementation in Go.
-- @param request the request of the connection
function canHandle(request)

end

--- Define the pre handle for the request
-- preHandle function is called by the desired service and pre handles the request by the connection. This function can
-- be used to alter, log or visualize the incoming request in the Honeytrap. From the Honeytrap implementation in Go
-- there are protocol or service specific functions declared. Take a look in the Honeytrap documentation to check the
-- availability for standard services. When implementing your own service you can freely implement the preHandle
-- function in your service.
-- @param message command of the request
function preHandle(message)
    return message
end

--- Define the handle of the request
-- handle function is called by the desired service and handles the request by the connection. From the Honeytrap
-- implementation in Go there are protocol or service specific functions declared. Take a look in the Honeytrap
-- documentation to check the availability for standard services. When implementing your own service you can freely
-- implement the handle function in your service.
-- @param message command of the request
function handle(message)
    return "Hello Lua! Your message:"..message..", was received from ".. getRemoteAddr() .." on ".. getDatetime() .."!"
end

--- Define the after handle of the request
-- afterHandle function is called by the desired service and after handles the request by the connection. This function
-- can be used to alter, log or visualize the outgoing response in the Honeytrap. From the Honeytrap implementation in
-- Go there are protocol or service specific functions declared. Take a look in the Honeytrap documentation to check the
-- availability for standard services. When implementing your own service you can freely implement the afterHandle
-- function in your service.
-- @param message response of the request
function afterHandle(message)
    return message
end
