// @flow
import React, {Component} from 'react'
import type {Props} from './render'
import {Box, Text, Icon} from '../common-adapters'
import type {Props as IconProps} from '../common-adapters/icon'
import Row from './row'
import {globalStyles, globalColors} from '../styles/style-guide'

type State = {
  showIgnored: boolean
}

const Ignored = ({showIgnored, ignored, isPublic, onToggle}) => {
  const topBoxStyles = {
    backgroundColor: isPublic ? globalColors.lightGrey : globalColors.darkBlue3,
    color: isPublic ? globalColors.black_40 : globalColors.white_75,
    borderTop: isPublic ? 'solid 1px rgba(0, 0, 0, 0.05)' : 'solid 1px rgba(255, 255, 255, 0.05)'
  }

  const bottomBoxStyles = {
    backgroundColor: isPublic ? globalColors.lightGrey : globalColors.darkBlue3,
    color: isPublic ? globalColors.black_40 : globalColors.white_40
  }

  const icon: IconProps.type = `caret-${showIgnored ? 'down' : 'right'}-${isPublic ? 'black' : 'white'}`

  return (
    <Box style={stylesIgnoreContainer}>
      <Box style={{...topBoxStyles, ...stylesIgnoreDivider}} onClick={onToggle}>
        <Text type='BodySmallSemibold' style={stylesDividerText}>Ignored folders</Text>
        <Icon type={icon} style={stylesIgnoreCaret} />
      </Box>
      {showIgnored && <Box style={{...bottomBoxStyles, ...stylesIgnoreDesc}}>
        <Text type='BodySmallSemibold' style={stylesDividerBodyText}>Ignored folders won't show up on your computer and you won't receive alerts about them.</Text>
      </Box>}
      {showIgnored && ignored.map((i, idx) => (
        <Row
          key={i.users.map(u => u.username).join('-')}
          users={i.users}
          isPublic={isPublic}
          ignored
          isFirst={!idx} />
        ))}
    </Box>
  )
}

class Render extends Component<void, Props, State> {
  state: State;

  constructor (props: Props) {
    super(props)

    this.state = {
      showIgnored: false
    }
  }

  render () {
    const realCSS = `
      .folder-row .folder-row-hover-action { visibility: hidden }
      .folder-row:hover .folder-row-hover-action { visibility: visible }
    `

    return (
      <Box style={stylesContainer}>
        <style>{realCSS}</style>
        {this.props.tlfs && this.props.tlfs.map((t, idx) => (
          <Row
            key={t.users.map(u => u.username).join('-')} {...t}
            isPublic={this.props.isPublic}
            ignored={false}
            isFirst={!idx} />
          ))}
        <Ignored ignored={this.props.ignored} showIgnored={this.state.showIgnored} isPublic={this.props.isPublic}
          onToggle={() => this.setState({showIgnored: !this.state.showIgnored})} />
      </Box>
    )
  }
}

const stylesContainer = {
  ...globalStyles.flexBoxColumn,
  flex: 1
}

const stylesIgnoreContainer = {
  ...globalStyles.flexBoxColumn
}

const stylesIgnoreDesc = {
  ...globalStyles.flexBoxColumn,
  alignItems: 'center'
}

const stylesIgnoreDivider = {
  ...globalStyles.flexBoxRow,
  alignItems: 'center',
  padding: 7,
  height: 32
}

const stylesDividerText = {
  ...globalStyles.clickable,
  color: 'inherit',
  marginRight: 7
}

const stylesDividerBodyText = {
  width: 360,
  padding: 7,
  textAlign: 'center',
  color: 'inherit'
}

const stylesIgnoreCaret = {
  color: globalColors.white_75
}

export default Render
