import UIKit

func appDelegate() -> AppDelegate {
  return UIApplication.sharedApplication().delegate as! AppDelegate
}

@UIApplicationMain
@objc(AppDelegate)
class AppDelegate: UIResponder {
  
  var window: UIWindow?
  var engine: Engine!
  
  private func setupReactWithOptions(launchOptions: [NSObject: AnyObject]?) -> RCTRootView {
    return RCTRootView(bundleURL: {
      
      #if DEBUG
        if let reactHost = AppDefault.ReactHost.stringValue {
          return NSURL(string: "http://\(reactHost)/react/index.bundle?platform=ios&dev=true")
        } else {
          return NSBundle.mainBundle().URLForResource("main", withExtension: "jsbundle")
        }
      #else
        return NSBundle.mainBundle().URLForResource("main", withExtension: "jsbundle")
      #endif
    }(), moduleName: "Keybase", initialProperties: nil, launchOptions: launchOptions)
  }
  
  private func setupEngine() {
    #if SIMULATOR
      let SecurityAccessGroupOverride = true
    #else
      let SecurityAccessGroupOverride = false
    #endif

    var home = AppDefault.HomeDirectory.stringValue ?? ""
    if home == "" {
      home = NSHomeDirectory()
    } else {
      let root = (NSHomeDirectory() as NSString).stringByAppendingPathComponent("Library")
      home = (root as NSString).stringByAppendingPathComponent(home)
    }

    engine = Engine(settings: [
      "runmode": AppDefault.RunMode.stringValue!,
      "homedir": home,
      "serverURI": AppDefault.APIServer.stringValue ?? "",
      "SecurityAccessGroupOverride": SecurityAccessGroupOverride
    ])
  }
  
}

class KeyListener: UIViewController {
  override func canBecomeFirstResponder() -> Bool {
    return true
  }

  var bridge: RCTBridge!

  override var keyCommands: [UIKeyCommand]? {
    return [
      UIKeyCommand(input: "[", modifierFlags: .Command, action: "goBackInTime:"),
      UIKeyCommand(input: "]", modifierFlags: .Command, action: "goForwardInTime:"),
      UIKeyCommand(input: "s", modifierFlags: [.Shift, .Command], action: "saveState:"),
      UIKeyCommand(input: "c", modifierFlags: [.Shift, .Command], action: "clearState:")
    ]
  }

  func goBackInTime(sender: UIKeyCommand){
    bridge.eventDispatcher.sendAppEventWithName("backInTime", body: true)
  }

  func goForwardInTime(sender: UIKeyCommand){
    bridge.eventDispatcher.sendAppEventWithName("forwardInTime", body: true)
  }

  func saveState(sender: UIKeyCommand){
    bridge.eventDispatcher.sendAppEventWithName("saveState", body: true)
  }

  func clearState(sender: UIKeyCommand){
    bridge.eventDispatcher.sendAppEventWithName("clearState", body: true)
  }

}

extension AppDelegate: UIApplicationDelegate {

  func application(application: UIApplication, didFinishLaunchingWithOptions launchOptions: [NSObject : AnyObject]?) -> Bool {
    #if DEBUG
      AppDefault.RunMode.setDefaultValue("devel")
    #else
      AppDefault.RunMode.setDefaultValue("prod")
    #endif
    
    #if SIMULATOR
      AppDefault.ReactHost.setDefaultValue("localhost:8081")
    #else
      #if DEBUG
        // Uncomment if you want your device to hit a local server while debugging
//        AppDefault.ReactHost.setDefaultValue("192.168.1.50:8081")
//        AppDefault.APIServer.setDefaultValue("http://192.168.1.50:3000")
      #endif
    #endif

    setupEngine()
    
    let rootViewController = KeyListener()
    let rctView = setupReactWithOptions(launchOptions)
    rootViewController.view = rctView
    rootViewController.bridge = rctView.bridge
    
    let window = UIWindow(frame: UIScreen.mainScreen().bounds)
    self.window = window
    window.rootViewController = rootViewController
    window.makeKeyAndVisible()
    return true
  }
  
}