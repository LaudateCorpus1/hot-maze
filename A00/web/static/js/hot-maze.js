
// Those are a few hundred lines of highly questionable JS.
// They are not representative of state-of-the-art frontend practices.


// Workflow A00: put the data in the QR-code. No cloud.
//
// 1) The page displays a File input/Drag'n'drop zone
// 2) The User drops a file F
// 4) The page displays a QR-code containing the data of F
// 5) The user scans the QR-code with their mobile
// 6) The mobile downloads F

let resourceFile;
let qrText;

function showError(errmsg) {
  console.log("Error: " + errmsg)
  clearQR();
  freezeResize();
  errorZone.innerHTML = errmsg;
  errorZone.style.display = "block";
  return false;
}

let qrHundredth = 40;

function render(colorDark, clickCallback){
  // qrHundredth == 95 is a "big QR-code" (almost fullscreen)
  // qrHundredth == 50 is a "medium-size QR-code"

  // -50 is for the small [?] box on the left
  var w = Math.max(document.documentElement.clientWidth -50, (window.innerWidth || 0) -50) * qrHundredth / 100;
  var h = Math.max(document.documentElement.clientHeight, window.innerHeight || 0) * qrHundredth / 100;
  var c = 400;
  if ( w>0 )
    c = w;
  if ( h>0 && h<c )
    c = h;

  clearQR();
  console.log("Rendering QR data ", qrText);
  console.log(`${qrText.length} characters`);
  // The limit seems to be 2328 characters.
  // If the qrText exceeds, we get TypeError: Cannot read property '1' of undefined
  try {
    new QRCode("qrcode", {
      text: qrText,
      width: c,
      height: c,
      colorDark : colorDark,
      colorLight : "white",
      correctLevel : QRCode.CorrectLevel.M
    });
  } catch(e) {
    showError("Could not render QR-code: " + e);
    return;
  }

  // The QR-code contents is clickable to embiggen, not the whole qr-zone
  if (clickCallback)
    for(var i=0; i < qrcode.childNodes.length; i++)
      qrcode.childNodes[i].onclick = clickCallback;
}

// Go fullscreen -> bigger QR
var resizeTimeOut = null;
window.onresize = function(){
    if (resizeTimeOut != null)
        clearTimeout(resizeTimeOut);
    resizeTimeOut = setTimeout(function(){
        render("black", embiggen);
    }, 300);
};
function freezeResize(){
  // after scan, or on fatal error, we don't need/want to
  // redisplay a fresh black QR-code again.
  window.onresize = function(){ return false; }
}

function clearQR() {
  var node = qrcode;
  while (node.firstChild) {
      node.removeChild(node.firstChild);
  }
}

questionMark.onclick = function(event) { 
    if ( helpContents.style.display === "none" ){
      helpContents.style.display = "block";
      questionMark.style.display = "none";
    }
}
helpClose.onclick = function(event) { 
    helpContents.style.display = "none";
    questionMark.style.display = "";
}
document.onkeydown = function(evt) {
  evt = evt || window.event;
  var isEscape = false;
  if ("key" in evt) {
    isEscape = (evt.key == "Escape" || evt.key == "Esc");
  } else {
    isEscape = (evt.keyCode == 27);
  }
  if (isEscape) {
    questionMark.style.display = "";
    helpContents.style.display = "none";
  }
};

function embiggen(event) { 
    // Change size on each click
    // 40, 68, 96, 40, 60, 96 ...
    qrHundredth += 28;
    if ( qrHundredth>100 )
      qrHundredth = 40;
    render("black", embiggen);
}

window.addEventListener('offline', function(){
  showError("Lost internet connectivity :(")
});

function displayGetQrCode() {
  // We generate the QR-code of a custom "medium-size" shortURL,
  // much shorter than the full signed GCS download URL.

  resourceFile.text().then(text => {
    const limit = 2328; // bytes
    if ( text.length > limit ) {
      showError(`Can't encode ${resourceFile.size} bytes in QR-code, max is ${limit}`);
      return;
    }
    qrText = text;
    render("black", embiggen);
  });
}

function processResourceFile() {
   // Here, resourceFile has been set either through the file select input,
   // or through a drag'n'drop.
   // What happens next with resourceFile is the same in both use cases.

  formZone.style.display = "none";
   displayGetQrCode();
}

resourceInput.onchange = function(e) { 
    resourceFile = resourceInput.files[0];
    processResourceFile();
};

function handleDnd(){
  // See https://css-tricks.com/drag-and-drop-file-uploading/
  var div = document.createElement('div');
  var dndFileSupported = (('draggable' in div) 
    || ('ondragstart' in div && 'ondrop' in div)) 
    && 'FormData' in window && 'FileReader' in window;
  if (!dndFileSupported)
    return;

  dropInvite.style.display = "block";

  var dropZone = resourceUpload;

  function highlightDropZone(){
    dropZone.style.border = 'red dotted 10px';
  }

  function unhighlightDropZone(){
    dropZone.style.border = 'rgba(255, 255, 255, 0) solid 10px';
  }

  dropZone.addEventListener('dragover', function(evt){
    evt.stopPropagation();
    evt.preventDefault();
    highlightDropZone()
    evt.dataTransfer.dropEffect = 'copy';
  }, false);

  dropZone.addEventListener('drop', function(evt){
    evt.stopPropagation();
    evt.preventDefault();
    helpContents.style.display = "none";
    unhighlightDropZone();

    var files = evt.dataTransfer.files;
    if( files.length>1 ){
      showError("Multiple upload not supported (yet)");
      return;
    }

    // Upload 1st file only.
    resourceFile = files[0];
    processResourceFile();
  }, false);

  dropZone.addEventListener('dragleave', function(evt){
    unhighlightDropZone();
  }, false);
}

//
// Let's start: init, then wait for user file selection.
//

handleDnd();
