//
//  Settings.m
//  Keybase
//
//  Created by Gabriel on 1/11/16.
//  Copyright © 2016 Keybase. All rights reserved.
//

#import "Settings.h"

@interface Settings ()
@property NSString *appPath;
@property NSString *runMode;
@property GBSettings *settings;
@end

@implementation Settings

- (instancetype)init {
  if ((self = [self initWithSettings:[GBSettings settingsWithName:@"Settings" parent:nil]])) {
  }
  return self;
}

- (instancetype)initWithSettings:(GBSettings *)settings {
  if ((self = [super init])) {
    NSArray *args = NSProcessInfo.processInfo.arguments;
    self.settings = settings;
    GBCommandLineParser *parser = [[GBCommandLineParser alloc] init];
    [parser registerOption:@"app-path" shortcut:'a' requirement:GBValueRequired];
    [parser registerOption:@"run-mode" shortcut:'r' requirement:GBValueRequired];
    [parser registerSettings:self.settings];
    NSArray *subargs = [args subarrayWithRange:NSMakeRange(1, args.count-1)];
    [parser parseOptionsWithArguments:subargs commandLine:args[0]];
    self.runMode = [self.settings objectForKey:@"run-mode"];
    NSAssert(self.runMode, @"No run mode");
    self.appPath = [self.settings objectForKey:@"app-path"];
    NSAssert(self.appPath, @"No app path");
  }
  return self;
}

- (KBEnvironment *)environment {
  NSString *servicePath = [self.appPath stringByAppendingPathComponent:@"Contents/SharedSupport/bin"];
  KBEnvironment *environment = [KBEnvironment environmentForRunModeString:self.runMode servicePath:servicePath];
  return environment;
}

@end
