import React from 'react';
import './Table.css';

const Table = ({ data }) => {
return (
<table className="table">
<thead>
<tr>
<th>IP Address</th>
<th>Ping Time</th>
<th>Last Successful</th>
</tr>
</thead>
<tbody>
{data.map((result, index) => (
<tr key={index}>
<td>{result.ip_address}</td>
<td>{result.ping_time}</td>
<td>{result.last_successful || "N/A"}</td>
</tr>
))}
</tbody>
</table>
);
};

export default Table;
