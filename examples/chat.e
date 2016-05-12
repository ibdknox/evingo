build the chat pane
  #chat-pane channel
  add
    chat-pane
      #div
      class: "chat-pane"
      children:
        #div class: "chat-messages" id: "{channel}-chat-messages"
        #input keydown: "chat-input-keydown"
          info-chat-channel: channel

draw messages
  #message name time message channel
  add
    #div
      parent: "{chat-pane}-chat-messages"
      class: "chat-message"
      children:
        #div class: "chat-user" text: "{name}"
        #div class: "chat-message"    text: "{message}"

handle chat keydowns
  #chat-input-keydown value info-channel: channel
  #keyboard pressed: "enter"
  #user name
  add forever
    #message 
      name
      time: printTime 
      message: value 
      channel