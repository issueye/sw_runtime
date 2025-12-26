// utils.js - 工具模块
const path = require('path');

function formatMessage(message) {
    return `[${new Date().toISOString()}] ${message}`;
}

function isValidPath(filePath) {
    return typeof filePath === 'string' && filePath.length > 0;
}

function joinPaths(...paths) {
    return path.join(...paths);
}

module.exports = {
    formatMessage,
    isValidPath,
    joinPaths
};