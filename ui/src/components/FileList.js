import React from 'react';

//const listUrl = '/document-storage/json/2/docs';
const listUrl = '/ui/api/list';

export default class FileList extends React.Component {
    constructor(props) {
        super(props)
        this.state = {
            loading: true,
            documentList : [],
            lastError: null
        }
    }
    componentDidMount(){
        fetch(listUrl, {
            method: 'GET',
            headers: new Headers({
              'Authorization': 'some_token'
            })
        })
            .then(r => {
                if (!r.ok) {
                    throw Error(r.statusText)
                }
                return r
            })
            .then(r => r.json())
            .then(documentList => this.setState( { documentList, loading:false}))
            .catch(e => this.setState( { lastError : e, loading: false }))
    }

    render(){
        if (this.state.loading) {
            return <div>Loading...</div>
        }
        if (this.state.lastError) {
            return <div>{this.state.lastError.message}</div>
        }

        let n = this.state.documentList.length
        if (n === 0) {
            return <div>No documents</div>
        }

        return (
            <table>
            { 

                this.state.documentList.map(x => 
            <tr>
                <td>{x.ID}</td>
                <td>{x.VissibleName}</td>
                <td>{x.ParentId}</td>
            </tr>
            )
            }
            </table>
        )
    }
}
