import LogComp from '~/console/components/logger';
import axios from "axios";
import {useEffect} from "react";
import https from "https";

const App = () => {

    useEffect(() => {
        (async()=>{
            try{
                var axios1 =await axios({
                    url:'http://10.13.0.1:17171',

                });

                console.log((axios1))
            }catch(err){
                // console.log(err.message)
            }

        })()
    }, []);
  return (
    <div className="flex items-center justify-center h-screen w-screen">
     kk
    </div>
  );
};

const Logs = () => {
  return <App />;
};

export default Logs;
