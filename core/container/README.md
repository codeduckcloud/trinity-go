// Author: Daniel TAN
// Date: 2021-12-18 00:22:22
// LastEditors: Daniel TAN
// LastEditTime: 2021-12-18 00:23:29
// FilePath: /trinity-micro/core/ioc/container/README.go
// Description: 
# container
the container package provide the dependency injection     
## Overview
### NewContainer
new a container instance 
```
// new container with default config 
s := NewContainer()

// new container with customize config 
s := NewContainer(Config{
	// the default value of autowired 
    // default true
	AutoWire        bool
    // replace the log
    // default logrus.New()  
	Log              : logrus.New(),
    // replace the container keyword
    // default container  
	ContainerKeyword : Keyword("container"),  
    // replace the autowire keyword
    // default autowire  
	AutowireKeyword  : Keyword("autowire"),  
    // replace the resource keyword
    // default resource 
	ResourceKeyword  : Keyword("resource"),   
})
```

### RegisterInstance (Singleton)
register a singleton instance.
#### `Only can register one kinds of instance, singleton or multi-instance`
```
instance1:= UserService{}
s := NewContainer()
s.RegisterInstance("instance1", &instance1)
```


### RegisterInstance (Multi-instance)
register a multi-instance instance.
#### `Only can register one kinds of instance, singleton or multi-instance`
```
instancePool1:= Sync.Pool{
    New: func() interface{}{
        return new(UserService)
    }
}
s := NewContainer()
s.RegisterInstancePool("instance1", instancePool1)
```

### InstanceDISelfCheck 
do the di self check 
#### `if the registered instance's param with autoware not be injected, will return error`
-  success case 
```
type UserService struct {
    UserRepo UserRepo `container:"autowire:true;resource:UserRepo"`
}

UserServicePool:= Sync.Pool{
    New: func() interface{}{
        return new(UserService)
    }
}
UserRepoPool:= Sync.Pool{
    New: func() interface{}{
        return new(UserRepo)
    }
}
s := NewContainer()
s.RegisterInstancePool("UserService", UserServicePool)
s.RegisterInstancePool("UserRepo", UserRepoPool)
err := s.InstanceDISelfCheck() // err is nil 
```
- failed case       
```
type UserService struct {
    UserRepo UserRepo `container:"autowire:true;resource:UserRepo"`
}

UserServicePool:= Sync.Pool{
    New: func() interface{}{
        return new(UserService)
    }
}
s := NewContainer()
s.RegisterInstancePool("UserService", UserServicePool)
err := s.InstanceDISelfCheck() // err is not nil
// cause autowire is true, but there is no instance "UserRepo" registered in container
```


### GetInstance 
get an instance from container
- normal case 
```
UserServicePool:= Sync.Pool{
    New: func() interface{}{
        return new(UserService)
    }
}
s := NewContainer()
s.RegisterInstancePool("UserService", UserServicePool)
UserService := s.GetInstance("UserService")
```
- inject with existing instance
```
UserServicePool:= Sync.Pool{
    New: func() interface{}{
        return new(UserService)
    }
}
s := NewContainer()
s.RegisterInstancePool("UserService", UserServicePool)
// used to handle the Circular References.
// always inject with the existing instance as priority
UserService := s.GetInstance("UserService", map[string]interface{}{
    "UserRepo":&UserRepo{},
})
```

### Release (multi instance)
release the instance to the pool and clean all the param which has been injected
only effective in multi-instance 
```
type UserService struct {
    UserRepo UserRepo `container:"autowire:true;resource:UserRepo"`
}

UserServicePool:= Sync.Pool{
    New: func() interface{}{
        return new(UserService)
    }
}
UserRepoPool:= Sync.Pool{
    New: func() interface{}{
        return new(UserRepo)
    }
}
s := NewContainer()
s.RegisterInstancePool("UserService", UserServicePool)
s.RegisterInstancePool("UserRepo", UserRepoPool)
UserService := s.GetInstance("UserService")
s.Release("UserService",UserService)
// s.Release("UserRepo",UserService.UserRepo ) 
// s.Release("UserService",UserService ) 
```

### Release (multi instance)

## Getting Started
- How to use Container pkg 
```
s := NewContainer()
s.RegisterInstancePool("UserService", UserServicePool)
s.RegisterInstancePool("UserRepo", UserRepoPool)
if err := s.InstanceDISelfCheck; err != nil {
    log.Fatal("di self check failed")
}
instance := s.GetInstance("instance")
defer s.Release("instance", instance)
// service logic 
instance.XXX 

```
## Contributing

Feel free to create the Issue and PR . We need your help !