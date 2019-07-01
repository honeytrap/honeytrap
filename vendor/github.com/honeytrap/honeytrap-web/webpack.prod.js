const webpack = require('webpack');

const { gitDescribeSync } = require('git-describe');
const gitInfo = gitDescribeSync();

const common = require('./webpack.common.js');
const merge = require('webpack-merge');

const UglifyJSPlugin = require('uglifyjs-webpack-plugin');

module.exports = merge(common, {
    devtool: 'cheap-module-source-map',
    plugins: [
        new webpack.DefinePlugin({
            "process.env": {
                NODE_ENV: JSON.stringify("production"),
                WEBSOCKET_URI: null,
                CLIENT_VERSION: JSON.stringify(gitInfo.raw)
            }
        }),
        new UglifyJSPlugin(),
    ]
});

