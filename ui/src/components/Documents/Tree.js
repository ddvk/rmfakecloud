import { React, useEffect, useState } from 'react';
import { Treebeard } from 'react-treebeard';
import apiservice from "../../services/api.service"

//TODO: poc only
const treeStyle = {
    tree: {
        base: {
            listStyle: 'none',
            // backgroundColor: 'white',
            padding: 0,

            // color: 'white',
            fontFamily: '"Helvetica Neue", "Open Sans", Arial, sans-serif',
            fontSize: '1.3rem'
        },
        node: {
            base: {
                position: 'relative',
                border:1,
                borderStyle:'solid',
                borderColor:'black',
                margin:5
            },
            activeLink: {
                background: 'blue'

            },
            toggle: {
                base: {
                    position: 'relative',
                    display: 'inline-block',
                    verticalAlign: 'top',
                    marginLeft: '5px',
                    height: '13px',
                    width: '14px'
                },
                wrapper: {
                    // position: 'absolute',
                    verticalAlign: 'bottom',
                    // top: '50%',
                    // left: '50%',
                    // margin: '-7px 0 0 -7px',
                    margin: '0px 0px 0 0',
                    height: '14px'
                },
                height: 14,
                width: 14,
                arrow: {
                    // fill: 'rgb(35,31,32)',
                    strokeWidth: 1
                }
            },
            header: {
                base: {
                    display: 'inline-block',
                    verticalAlign: 'top',
                    // color: 'rgb(35,31,32)'
                },
                connector: {
                    width: '2px',
                    height: '12px',
                    // borderLeft: 'solid 2px black',
                    // borderBottom: 'solid 2px black',
                    position: 'absolute',
                    // top: '0px',
                    // left: '-21px'
                },
                title: {
                    lineHeight: '24px',
                    // verticalAlign: 'middle',
                    padding: '5px'
                }
            },
            subtree: {
                listStyle: 'none',
                border:2,
                borderColor: 'red',
                paddingLeft: '19px'
            },
            loading: {
                color: '#E2C089'
            }
        }
    }
}
const TreeExample = (props) => {
    const [data, setData] = useState({
        loading: true
    });
    const [downloadError, setDownloadError] = useState();
    const [dwn, setDownloadUrl] = useState();
    const [cursor, setCursor] = useState(false);
    const {onFileSelected} = props

    const loadDocs = () =>
        apiservice.listDocument()
        .then(docs => {
            setData({
                loading:false,
                docs: docs.Entries
            })
        })
        .catch(e => {
            setData({
                loading: false,
                error: e
            })
        })

    useEffect(() => {
        loadDocs()
    },[props.counter])

    useEffect(()=>{
        document.addEventListener("contextmenu", (event)=>{
        })
    });

    if (data.loading) {
        return <div>Loading...</div>;
    }
    if (data.error) {
        return <div>{data.error.message}</div>;
    }

    if (data.docs && !data.docs.length) {
        return <div>No documents</div>;
    }
    const onToggle = (node, toggled) => {
        if (cursor) {
            cursor.active = false;
        }
        node.active = true;
         console.log(node.id)
        if (node.children) {
            node.toggled = toggled;
            setDownloadUrl(null);
            if (onFileSelected) {
                onFileSelected(null);
            }
            props.onFolderChanged(node.id);
        } else {
            //TODO: another quick poc hack
            setDownloadUrl({id:node.id, name:node.name})
            if (onFileSelected) {
                onFileSelected(node.id);
            }
        }
        setCursor(node);
        setData(Object.assign({}, data))
    }

    const onDownloadClick = () => {
        setDownloadError(null)
        const {id, name} = dwn
        apiservice.download(id)
        .then(blob => {
            var url = window.URL.createObjectURL(blob)
            var a = document.createElement('a')
            a.href = url
            a.download = name+ '.pdf'
            document.body.appendChild(a)
            a.click()
            a.remove()
        })
        .catch(e => {
            setDownloadError('cant download ' + e)
        })

    }

    return (
        <div style={{"marginTop":"20px"}}>
            { dwn && <button onClick={onDownloadClick}>Download {dwn.name}</button> }
            { downloadError && <div class="error">{downloadError}</div> }
            <Treebeard style={treeStyle} data={data.docs} animations={false} onToggle={onToggle} />
        </div>
    )
}
export default TreeExample;
