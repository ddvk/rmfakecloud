import React from "react";

import useFetch from "../../hooks/useFetch";

export default function List() {
  //const skuRef = useRef();
  //const { id } = useParams();
  //const navigate = useNavigate();

  const { data: documentList, error, loading } = useFetch('list');

  if (loading) {
    return <div>Loading...</div>;
  }
  if (error) {
    return <div>{error.message}</div>;
  }

  if (!documentList.Entries.length) {
    return <div>No documents</div>;
  }

  return (
    <div>
      {documentList.Entries.map((x) => (
        <div key={x.id} className="documentEntry">
          <p>{x.name}</p>
        </div>
      ))}
    </div>
  );
}
