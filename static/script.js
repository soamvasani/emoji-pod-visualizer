'use strict';

// Create a new HTML5 EventSource
var source = new EventSource('/events/');

// Create a callback for when a new message is received.
source.onmessage = function(e) {
    console.log("got: ", e.data);
    var o = JSON.parse(e.data);
    handleEvent(o);
};


var ps = [];

function handleEvent(e) {
    //e.podName
    //e.containerStates : name->state

    ps[e.podName] = e.containerStates;

    redraw();
}

const running = "";
const creating = "";
const error = "";

function redraw() {
    var out = "";
    for (var podName in ps) {
        //out += JSON.stringify(ps[podName]);
        //out += "<br/>";
        var cStates = ps[podName];
        
        out += '<div class="pod">';
        out += '<span class="podName">' + podName + "</span>";
        for (var cName in cStates) {
            var state = cStates[cName];
            if (state === 'running') {
                out += '<img src="/static/smiley.png">';
            } else if (state === 'ContainerCreating') {
                out += '<img src="/static/baby.png">';
            } else if (state === 'terminated') {
                out += '<img src="/static/ghost.png">';
            } else if (state === 'ErrImagePull') {
                out += '<img src="/static/unamused.png">';
            } else if (state === 'ImagePullBackOff') {
                out += '<img src="/static/frown.png">';
            } else {
                out += '<img src="/static/question-mark-emoji.png">';
            }
        }
        out += '</div>';
        
    }
    $("#podinfo").html(out);
}

// // Create a new HTML5 EventSource
// var source = new EventSource('/events/');

// // Create a callback for when a new message is received.
// source.onmessage = function(e) {
//     console.log("got: ", e.data);
//     var o = JSON.parse(e.data);
//     handleEvent(o);
// };


// var ps = [];

// function handleEvent(e) {
//     //e.podName
//     //e.containerStates : name->state

//     ps[e.podName] = e.containerStates;

//     redraw();
// }

// function redraw() {
//     var out = "";
//     for (var podName in ps) {
//         out += JSON.stringify(ps[podName]);
//         out += "<br/>";
//     }
//     $("#podinfo").html(out);
//}
