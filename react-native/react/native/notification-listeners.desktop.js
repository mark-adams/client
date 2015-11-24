import notify from '../../../desktop/app/hidden-window-notifications'
import enums from '../../react/constants/types/keybase_v1'
import type {FSNotification} from '../../react/constants/types/flow-types'

export default {
  'keybase.1.NotifySession.loggedOut': () => {
    notify('Logged Out')
  },
  'keybase.1.NotifyFS.FSActivity': params => {
    const notification: FSNotification = params.notification

    const action = {
      [enums.kbfs.FSNotificationType.encrypting]: 'Encrypting and uploading',
      [enums.kbfs.FSNotificationType.decrypting]: 'Downloading, decrypting and verifying',
      [enums.kbfs.FSNotificationType.signing]: 'Signing and uploading',
      [enums.kbfs.FSNotificationType.verifying]: 'Downloading and verifying',
      [enums.kbfs.FSNotificationType.rekeying]: 'Rekeying'
    }[notification.notificationType]

    const state = {
      [enums.kbfs.FSStatusCode.start]: '',
      [enums.kbfs.FSStatusCode.finish]: 'finished',
      [enums.kbfs.FSStatusCode.error]: 'errored'
    }[notification.statusCode]

    const pubPriv = notification.publicTopLevelFolder ? '[Public]' : '[Private]'

    const title = `KBFS: ${action} ${state}`
    const body = `File: ${notification.filename} ${pubPriv} ${notification.status}`

    notify(title, {body})
  }
}
