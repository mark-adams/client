// Copyright 2004-present Facebook. All Rights Reserved.

/**
 * React Native CLI configuration file
 */
'use strict'

const blacklist = require('react-native/packager/blacklist.js')
const path = require('path')

module.exports = {
  getProjectRoots() {
    return this._getRoots()
  },

  getAssetRoots() {
    return this._getRoots()
  },

  getBlacklistRE() {
    return blacklist('')
  },

  _getRoots() {
    return [
      path.resolve(__dirname, '.'),
      path.resolve(__dirname, '..', 'react'),
      path.resolve(__dirname, 'node_modules')];
    /*
    // match on either path separator
    if (__dirname.match(/node_modules[\/\\]react-native[\/\\]packager$/)) {
      // packager is running from node_modules of another project
      return [path.resolve(__dirname, '../../..')];
    } else if (__dirname.match(/Pods\/React\/packager$/)) {
      // packager is running from node_modules of another project
      return [path.resolve(__dirname, '../../..')];
    } else {
      return [path.resolve(__dirname, '..')];
    }
    */
  }
}
