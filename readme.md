# Value Expressions

```
nil
true
false
12
-10
33.29
-12e6
"this is a string"
7a78b9b7-7b04-4e5c-a57c-b17895667fd9
```


# Set Syntax

```
/dir/path(-10,false,(nil,12)) = "hello"

/deep/dir/path("years",true,10e3) = nil

/dir(12,"lives") = (true,false,false)
```


# Get Syntax

```
/dir/{}(1,2,3) = nil

/deep/dir/path({},true,10e3) = {}
```


# Clear Syntax

```
/dir/path(-10,false,(nil,12)) = clear
```
