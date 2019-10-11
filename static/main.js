

function getData(k){
    var xhttp = new XMLHttpRequest();
    xhttp.open("GET", "/get/"+k, false);
    xhttp.send();
    return JSON.parse(xhttp.responseText);
}

function setData(k,v){
    var xhttp = new XMLHttpRequest();
    xhttp.open("POST", "/set/"+k, true);
    xhttp.send(v);
    //return JSON.parse(xhttp.responseText);
}



const SortableList = {
  mixins: [ContainerMixin],
  template: `
    <ul class="list">
      <slot />
    </ul>
  `
};

const SortableItem = {
  mixins: [ElementMixin],
  props: ['header'],
  template: `
  <li class="list-item">
    <input class="hidden-box" type="checkbox" :id="header['name']+'-chk'" :value="header['name']" v-model="header['enabled']" v-on:change="updateHeader($event, header)" >
    <label class="check--label" :for="header['name']+'-chk'">
      <span class="check--label-box"></span>
      <span class="check--label-text">{{header['name']}}</span>
    </label>
  </li>
    `,
    methods:{
        updateHeader: function(e,h){
          //console.log(e,h,this);
          this.$emit('headerchange',[e,h])
          /*
        e.header['enabled'] = !e.header['enabled'];
        alert(e);*/
        //setData()
      }
    }
};//mattia 103 - macroobject wword to web 2011

const TablehHeaders = {
  name: 'TableHeaders',
  template: `
    <div class="root">
      <SortableList lockAxis="y" v-model="headers" :headerchange="headerUpdate" @sort-end="orderUpdated" :lockAxis="'xy'" :distance=5  :pressTreshold=10 >
        <SortableItem v-for="(header, index) in headers" :index="index" :key="header['name']" :header="header" @headerchange="headerUpdate"/>
      </SortableList>
    </div>
  `,
  components: {
    SortableItem,
    SortableList
  },
  data() {
    return {
        //selectedTypes: getData("tableHeaders"),
        headers: function() {
            let currentHeaders = getData("tableHeaders")
            let allHeaders = Object.keys(getData("types")).filter((e) => {
                return currentHeaders.indexOf(e) < 0
            }, currentHeaders);
            let h = []
            currentHeaders.forEach(e => {
                h.push({"name":e, enabled: true})
            },h);
            allHeaders.forEach(e => {
                h.push({"name":e, enabled: false})
            },h );
            return h
        }()
    };
  },
  methods:{
        headerUpdate: function(e){
          //console.log("CATCHED!", e, this.headers)

          for (var i = 0; i < this.headers.length; i++){
            if (this.headers[i] == e[1].name) {
              this.headers[i].enabled = e[0].target.checked
              break;
            }
          }
          let enabledHeaders = this.headers.filter((e) => {return e.enabled}).map((v)=> {return v.name})
          //console.log(headers)
          setData("tableHeaders", JSON.stringify(enabledHeaders))
          
          //Checked on top
          let unchecked =  this.headers.filter((e)=>{return !e.enabled})
          
          this.headers = enabledHeaders.map((e)=>{return {name: e, enabled: true}}).concat(unchecked)
          

        },
        orderUpdated: function(){
          let headers = this.headers.filter((e) => {return e.enabled}).map((v)=> {return v.name})
          setData("tableHeaders", JSON.stringify(headers))
        }
  }
};

         

var app = new Vue({
    el: '#app',
    render: h => h(TablehHeaders)
})