// Proxy configuration file. See link for more information:
// https://github.com/angular/angular-cli/blob/master/docs/documentation/stories/proxy.md

const PROXY_CONFIG = [{
    context: [
        "/api/v1/host",
    ],
    target: "http://localhost:8080",
    secure: false
}, {
    context: [
        "/api/v1/exec",
    ],
    target: "http://localhost:8080",
    secure: false,
    ws: true
}]

module.exports = PROXY_CONFIG;