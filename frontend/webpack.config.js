const HtmlWebpackPlugin = require('html-webpack-plugin');
const path = require('path');

module.exports = {
    entry: './src/parken.js',
    output: {
        filename: 'parken.[contenthash].js',
        path: path.resolve(__dirname, '../dist/frontend'),
        clean: true
    },
    plugins: [
        new HtmlWebpackPlugin({
            template: "./src/index.html"
        })
    ],
    module: {
        rules: [
            {
                test: /\.css$/i,
                use: ['style-loader', 'css-loader']
            },
            {
                test: /\.png$/i,
                type: 'asset/resource'
            }
        ]
    }
}
