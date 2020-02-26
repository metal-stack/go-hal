# UUID

- https://en.wikipedia.org/wiki/Universally_unique_identifier#Encoding
- http://dnaeon.github.io/convert-big-endian-uuid-to-middle-endian/
- https://github.com/dnaeon/go-uuid-endianness
- https://github.com/Nordix/go-redfish
- https://github.com/stmcginnis/gofish

```python
import json, requests
url='https://192.168.0.5/redfish/v1/Systems/1'
userid='ADMIN'
password='ADMIN'
r = requests.get(url, auth=(userid, password), verify=False)
jsonData = r.json()
print (jsonData)
```