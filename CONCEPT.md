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