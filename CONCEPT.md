## Translang concept idea

Figma url to test on: https://www.figma.com/design/uT0nuPS2pCAEH8J7ZB0Oi8/Accessibility?node-id=90-2607&t=MxOwcuX0cxtxQEpy-4

Set env in Powershell...:
```powershell
$env:FIGMA_PAT='VALUE'
$env:OPENAI_API_KEY='VALUE'
```

### User flow MVP
- Copy a section link from figma a figma file containing the text that you want to create translations for
- Paste the link to the cli
- The program fetches the Text node and takes an image of the text container node
- A copy key is generated from the figma text
- The text content gets translated to X languages using AI
- The cli presents the image, the copy key and the translations

### Tech
- Golang, because I want to

### Debug GCC issue
https://github.com/golang/go/issues/59490#issuecomment-1581874198
```
that happen because you installed gcc through cygwin64
so, try to install MinGW.
download it from this link https://github.com/niXman/mingw-builds-binaries/releases
make sure to choose a compatible version if you have Windows 64 choose mingw64 and so
then extract it into Windows partition in my case was C:\ partition
then change the gcc path in environment variables in my case the path after the update was "C:\mingw64\bin"
then restart vs code if it opens then try the command again.

if it needs to CG_ENABELED=1
then open powershell as admin and write this command "go env -w CGO_ENABLED=1"
```

### DB Structure

translation -> id, figmaSourceUrl, contextImageUrl, createdAt, syncedAt => primary key id
translationNode -> tranlsation_id, figmaTextNodeID => derived key translation_id, figmaTextNodeID
translationNodeValue -> copy_key, translation_node_id, language, text, createdAt, syncedAt => primary key copy_key unique

#### Simpler db handling lib
For easier handling of struct parsing
https://github.com/jmoiron/sqlx