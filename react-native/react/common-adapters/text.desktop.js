import React, {Component} from '../base-react'
import {globalStyles, globalColors} from '../styles/style-guide'
import type {Props} from './text'

export default class Text extends Component {
  props: Props;

  render () {
    const typeStyle = {
      'Header': styles.textHeader,
      'Body': styles.textBody,
      'Error': styles.textError
    }[this.props.type]

    const style = {
      ...typeStyle,
      ...(this.props.link ? styles.textLinkMixin : {}),
      ...(this.props.small ? styles.textSmallMixin : {}),
      ...(this.props.reversed ? styles.textReversedMixin : {}),
      ...(this.props.onClick ? globalStyles.clickable : {}),
      ...this.props.style
    }

    return <p className={this.props.link ? 'hover-underline' : ''} style={style} onClick={this.props.onClick}>{this.props.children}</p>
  }
}

Text.propTypes = {
  type: React.PropTypes.oneOf(['Header', 'Body']),
  link: React.PropTypes.bool,
  small: React.PropTypes.bool,
  reversed: React.PropTypes.bool,
  children: React.PropTypes.node,
  style: React.PropTypes.object,
  onClick: React.PropTypes.func
}

const textCommon = {
  ...globalStyles.fontRegular,
  ...globalStyles.noSelect,
  color: globalColors.grey1,
  cursor: 'default'
}

export const styles = {
  textHeader: {
    ...textCommon,
    ...globalStyles.fontBold,
    fontSize: 18,
    lineHeight: '22px',
    letterSpacing: '0.5px'
  },
  textBody: {
    ...textCommon,
    fontSize: 15,
    lineHeight: '20px',
    letterSpacing: '0.2px'
  },
  textError: {
    ...textCommon,
    color: globalColors.highRiskWarning,
    fontSize: 13,
    lineHeight: '17px',
    letterSpacing: '0.2px'
  },
  textLinkMixin: {
    color: globalColors.blue,
    cursor: 'pointer'
  },
  textSmallMixin: {
    color: globalColors.grey2,
    fontSize: 13,
    lineHeight: '17px'
  },
  textReversedMixin: {
    color: globalColors.white
  }
}
