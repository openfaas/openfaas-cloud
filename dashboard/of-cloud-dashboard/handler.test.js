const handlerFuncs = require('./handler.js')

test(`"gateway_url" contains the "dns_suffix" and shows backward compatibility for kubernetes users`, () => {
    expect(handlerFuncs.dnsFunc('http://gateway.openfaas:8080', 'openfaas')).toEqual('http://gateway.openfaas:8080')
})

test(`"gateway_url" is set like the old standard and shows backward compatibility for kubernetes users`, () => {
    expect(handlerFuncs.dnsFunc('http://gateway.openfaas:8080', '')).toEqual('http://gateway.openfaas:8080')
})

test(`"gateway_url" and "dns_suffix" are configured for kubernetes with new change`, () => {
    expect(handlerFuncs.dnsFunc('http://gateway:8080', 'openfaas')).toEqual('http://gateway.openfaas:8080')
})

test(`"gateway_url" is set and "dns_suffix" is unset which is the new configuration for Swarm`, () => {
    expect(handlerFuncs.dnsFunc('http://gateway:8080', '')).toEqual('http://gateway:8080')
})

test(`"gateway_url" is empty and goes back to the default "http://gateway:8080" with "dns_suffix" set`, () => {
    expect(handlerFuncs.dnsFunc('', 'openfaas')).toEqual('http://gateway.openfaas:8080')
})

test(`"gateway_url" is empty along with "dns_suffix" which shows that this is valid configuration for Swarm`, () => {
    expect(handlerFuncs.dnsFunc('', '')).toEqual('http://gateway:8080')
})

test(`"gateway_url" is set to a random string, but "dns_suffix" is unset`, () => {
    expect(handlerFuncs.dnsFunc('random_string', '')).toEqual('random_string')
})

test(`"gateway_url" and "dns_suffix" are set to a random strings which shows they are separated with dot`, () => {
    expect(handlerFuncs.dnsFunc('random_string', 'random_suffix')).toEqual('random_string.random_suffix')
})