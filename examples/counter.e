build the counter
  #counter count parent
  add
    #div class: "counter-container", parent
      children:
        #div #decrement class: "button", text: "-", diff: -1
        #div            class: "count",  text: "{count}"
        #div #increment class: "button", text: "+", diff: 1
        
increment the counter
  #click element
  #counter count
  choose 
    #increment element diff
  or
    #decrement element diff
  update forever 
    counter
      count: count + diff

go!
  add
    #counter count: 0, parent: "root"