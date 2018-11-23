# Goju

Goju is for _Go JSON UNIT_ tests.

## The idea

It is intended to test JSON files using other json files

```javascript
{
  "items": [ {
               "spec": {
                 "containers": [ {
                                 "image": "k8s.gcr.io/heapster-amd64:v1.5.0"
                                 }
                               ]
               }
             }
           ]
}
```

is checked for regular expression by the JSON file

```javascript
{
  "items": {
    "length" :"1",
    "spec":{
      "containers":{
        "image":{
           "matches":"^(gcr.io/(google[-_]containers|k8s-minikube)|k8s.gcr.io)"
        }
      }
    }
  }
}
```

This means, the _items_ array is checked for _length_ 1 and any images in the _items/spec/containers_ array is checked, if the string matches the regular expression.
More examples are in the data directory.

The concept is to check configurations by other configurations, and implement checks adding additional leaves in JSON or YAML.

The executable checks are invoked by reflection on a _Check_ object and must have the same name as define in the `check.go` file with a leading capital letter.

## Installation

Install _Goju_ by

```bash
go get github.com/endodoce/goju
```

## Usage

Simply call

```bash  
goju  -json=data/imagepod.json -rule=data/imagerule.json
```

to get an output like

```bash
I0624 13:24:16.733258    3384 main.go:29] Errors       : 0
I0624 13:24:16.733543    3384 main.go:30] Checks   true: 2
I0624 13:24:16.733573    3384 main.go:31] Checks  false: 0
```

The errors are parsing errors like unknown functions and the
check results, which are true and false are reported