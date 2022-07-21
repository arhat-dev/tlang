# Syntax of `tlang`

## Overview

Treated it as [golang template](https://pkg.go.dev/text/template) without `{{ }}`, pipelines are separated by new lines (and semi-colons).

## Comments

```tlang
# line comments only
# there is no plan for block comments support
```

## Text

```tlang
`line1`
"line2\n"
'\x00' # characters/runes
```

when evaluated, the above code generates `line1line2\n\0`

## Pipeline

```tlang
.X | doSomething | now
```

## Variables

## Control Flow

```tlang
range $_, $item := .List
  if .X
    break
  else if .Y
    continue
  else
    doSomething $item
  end
end
```

## Context Switching

```tlang
with .X
  doSomething .
end
```

## Template Definition

```tlang
define "hello"
  "Hallo"
end
```
