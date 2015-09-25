'use strict'
/* @flow */

import React from 'react-native'
const {
  Component,
  StyleSheet,
  View,
  Text
} = React

const commonStyles = require('../styles/common')

class About extends Component {
  constructor (props) {
    super(props)

    this.state = {}
  }

  // TODO get version from golang
  render () {
    return (
        <View style={styles.container}>
          <Text style={[{textAlign: 'center', marginBottom: 75}, commonStyles.h1]}>Version 0.1</Text>
        </View>
    )
  }

  static parseRoute (store, currentPath, nextPath) {
    const componentAtTop = {
      title: 'About',
      mapStateToProps: state => state.router.toObject(),
      component: About
    }

    return {
      componentAtTop,
      parseNextRoute: null
    }
  }
}

About.propTypes = {}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'stretch',
    backgroundColor: '#F5FCFF'
  }
})

module.exports = About