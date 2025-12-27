// logger.js - 日志模块，依赖 utils
const utils = require('./utils.js');

function log(message) {
    console.log(utils.formatMessage(`LOG: ${message}`));
}

function error(message) {
    console.error(utils.formatMessage(`ERROR: ${message}`));
}

function warn(message) {
    console.warn(utils.formatMessage(`WARN: ${message}`));
}

module.exports = {
    log,
    error,
    warn
};